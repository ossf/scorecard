package foo;

import java.lang.reflect.Field;

public class UnsafeFoo {
	public static void main(final String[] args) throws NoSuchFieldException, IllegalAccessException {
		final long address = getUnsafe().allocateMemory(0);
		for (final String s : args) {
			for (final char c : s.toCharArray()) {
				getUnsafe().putChar(address, c);
			}
		}
	}

	private static jdk.internal.misc.Unsafe getUnsafe() throws IllegalAccessException, NoSuchFieldException {
		final Field f = jdk.internal.misc.Unsafe.class.getDeclaredField("theUnsafe");
		f.setAccessible(true);
		return (jdk.internal.misc.Unsafe) f.get(null);
	}
}
