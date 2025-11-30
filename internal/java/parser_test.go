// Copyright 2025 OpenSSF Scorecard Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package java

import (
	"testing"
)

func TestParseFile_SunMiscUnsafe(t *testing.T) {
	t.Parallel()
	content := []byte(`package foo;

import sun.misc.Unsafe;

import java.lang.reflect.Field;

public class UnsafeFoo {
}
`)

	file, err := ParseFile(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should find sun.misc.Unsafe and java.lang.reflect.Field type names
	if len(file.TypeNames) < 2 {
		t.Fatalf("expected at least 2 type names, got %d", len(file.TypeNames))
	}

	// Check if sun.misc.Unsafe is found
	found := false
	for _, tn := range file.TypeNames {
		if tn.Name == "sun.misc.Unsafe" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected to find type name %q", "sun.misc.Unsafe")
	}
}

func TestParseFile_JdkInternalMiscUnsafe(t *testing.T) {
	t.Parallel()
	content := []byte(`package foo;

import jdk.internal.misc.Unsafe;

import java.lang.reflect.Field;

public class UnsafeFoo {
}
`)

	file, err := ParseFile(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should find jdk.internal.misc.Unsafe and java.lang.reflect.Field type names
	if len(file.TypeNames) < 2 {
		t.Fatalf("expected at least 2 type names, got %d", len(file.TypeNames))
	}

	// Check if jdk.internal.misc.Unsafe is found
	found := false
	for _, tn := range file.TypeNames {
		if tn.Name == "jdk.internal.misc.Unsafe" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected to find type name %q", "jdk.internal.misc.Unsafe")
	}
}

func TestParseFile_Malformed(t *testing.T) {
	t.Parallel()
	content := []byte(`
imp ort "sun.misc.Unsafe";

pub class SafeFoo {
`)

	file, err := ParseFile(content)
	// Should not return error even for malformed - parser is lenient
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should have no type names since the syntax is wrong
	if len(file.TypeNames) != 0 {
		t.Errorf("expected 0 type names for malformed file, got %d", len(file.TypeNames))
	}
}

func TestParseFile_UnsafeInMultipleContexts(t *testing.T) {
	t.Parallel()
	// Test with fully qualified names (no imports) to ensure type names are captured everywhere
	content := []byte(`package foo;

public class UnsafeFoo {
    private sun.misc.Unsafe unsafe;  // field declaration

    public sun.misc.Unsafe getUnsafe() { // return type
        sun.misc.Unsafe u = (sun.misc.Unsafe) sun.misc.Unsafe.getUnsafe();  // local variable, cast and static call
	return u;
    }

    public boolean doSomething(sun.misc.Unsafe param) {  // method parameter
        if (param instanceof sun.misc.Unsafe) { // instanceof
            return true;
        }
        if (sun.misc.Unsafe.class.equals(param.getClass())) { // class literal
            return true;
        }
        return false;
    }
}
`)

	file, err := ParseFile(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Count how many times sun.misc.Unsafe appears
	count := 0
	for _, tn := range file.TypeNames {
		t.Logf("Found type name: %q at line %d", tn.Name, tn.Pos().GetLine())
		if tn.Name == "sun.misc.Unsafe" {
			count++
		}
	}

	// Should find sun.misc.Unsafe in:
	// 1. field declaration (private sun.misc.Unsafe unsafe)
	// 2. return type (public sun.misc.Unsafe getUnsafe())
	// 3. local variable declaration (sun.misc.Unsafe u)
	// 4. cast expression ((sun.misc.Unsafe) null)
	// 5. static invocation (sun.misc.Unsafe.getUnsafe())
	// 6. method parameter (sun.misc.Unsafe param)
	// 7. instanceof operator (instanceof sun.misc.Unsafe)
	// 8. class literal (sun.misc.Unsafe.class)
	// Expect 8 occurrences
	if count != 8 {
		t.Errorf("expected 6 occurrences of sun.misc.Unsafe, got %d", count)
	}
}

func TestParseFile_IgnoreShortNames(t *testing.T) {
	t.Parallel()
	// Test that we capture fully qualified names but not short names (when imported)
	content := []byte(`package foo;

import sun.misc.Unsafe;

public class UnsafeFoo {
    private Unsafe unsafe;  // short name - not fully qualified

    public Unsafe getUnsafe() {  // short name - return type
        Unsafe u = null;  // short name - local variable
        return u;
    }

    public void doSomething(Unsafe param) {  // short name - parameter
        // method body
    }
}
`)

	file, err := ParseFile(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Count fully qualified vs short names
	fullyQualifiedCount := 0
	shortNameCount := 0
	for _, tn := range file.TypeNames {
		if tn.Name == "sun.misc.Unsafe" {
			fullyQualifiedCount++
		} else if tn.Name == "Unsafe" {
			shortNameCount++
		}
	}

	// Should only find sun.misc.Unsafe in the import statement
	if fullyQualifiedCount != 1 {
		t.Errorf("expected 1 fully qualified name (import), got %d", fullyQualifiedCount)
	}

	// Should find multiple short names (Unsafe) in field, return type, local var, parameter
	if shortNameCount < 4 {
		t.Errorf("expected at least 4 short names, got %d", shortNameCount)
	}
}
