// Copyright 2020 OpenSSF Scorecard Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package cmd implements Scorecard command-line.
package cmd

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"

	ngt "github.com/ossf/scorecard/v5/cmd/internal/nuget"
	pmc "github.com/ossf/scorecard/v5/cmd/internal/packagemanager"
)

func Test_fetchGitRepositoryFromNPM(t *testing.T) {
	t.Parallel()
	type args struct {
		packageName string
		result      string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "fetchGitRepositoryFromNPM",

			args: args{
				packageName: "npm-package",
				result: `
{
  "name": "@pulumi/pulumi",
  "version": "3.116.1",
  "description": "Pulumi's Node.js SDK",
  "license": "Apache-2.0",
  "repository": {
    "type": "git",
    "url": "git+https://github.com/pulumi/pulumi.git",
    "directory": "sdk/nodejs"
  },
  "dependencies": {
    "@grpc/grpc-js": "^1.10.1",
    "@logdna/tail-file": "^2.0.6",
    "@npmcli/arborist": "^7.3.1",
    "@opentelemetry/api": "^1.2.0",
    "@opentelemetry/exporter-zipkin": "^1.6.0",
    "@opentelemetry/instrumentation": "^0.32.0",
    "@opentelemetry/instrumentation-grpc": "^0.32.0",
    "@opentelemetry/resources": "^1.6.0",
    "@opentelemetry/sdk-trace-base": "^1.6.0",
    "@opentelemetry/sdk-trace-node": "^1.6.0",
    "@opentelemetry/semantic-conventions": "^1.6.0",
    "@pulumi/query": "^0.3.0",
    "@types/google-protobuf": "^3.15.5",
    "@types/semver": "^7.5.6",
    "@types/tmp": "^0.2.6",
    "execa": "^5.1.0",
    "fdir": "^6.1.1",
    "google-protobuf": "^3.5.0",
    "got": "^11.8.6",
    "ini": "^2.0.0",
    "js-yaml": "^3.14.0",
    "minimist": "^1.2.6",
    "normalize-package-data": "^6.0.0",
    "picomatch": "^3.0.1",
    "pkg-dir": "^7.0.0",
    "require-from-string": "^2.0.1",
    "semver": "^7.5.2",
    "source-map-support": "^0.5.6",
    "tmp": "^0.2.1",
    "upath": "^1.1.0"
  },
  "devDependencies": {
    "@types/ini": "^1.3.31",
    "@types/js-yaml": "^3.12.5",
    "@types/minimist": "^1.2.0",
    "@types/mocha": "^2.2.42",
    "@types/node": "14.18.3",
    "@types/normalize-package-data": "^2.4.0",
    "@types/picomatch": "^2.3.3",
    "@types/sinon": "^10.0.13",
    "@types/split2": "^2.1.6",
    "@types/uuid": "^9.0.7",
    "@typescript-eslint/eslint-plugin": "^4.29.0",
    "@typescript-eslint/parser": "^4.29.0",
    "eslint": "^7.32.0",
    "eslint-plugin-header": "^3.1.1",
    "eslint-plugin-import": "^2.23.4",
    "mocha": "^9.0.0",
    "nyc": "^15.1.0",
    "rome": "^12.1.0",
    "sinon": "^14.0.0",
    "ts-node": "^7.0.1",
    "typescript": "~3.8.3"
  },
  "peerDependencies": {
    "ts-node": "\u003E= 7.0.1 \u003C 12",
    "typescript": "\u003E= 3.8.3 \u003C 6"
  },
  "peerDependenciesMeta": {
    "typescript": {
      "optional": true
    },
    "ts-node": {
      "optional": true
    }
  },
  "pulumi": {
    "comment": "Do not remove. Marks this as as a deployment-time-only package"
  },
  "engines": {
    "node": "\u003E=18"
  },
  "mocha": {
    "require": [
      "ts-node/register",
      "source-map-support/register"
    ]
  },
  "//": [
    "NOTE: @types/node is pinned due to grpc/grpc-node#2002"
  ],
  "types": "./index.d.ts",
  "bugs": {
    "url": "https://github.com/pulumi/pulumi/issues"
  },
  "homepage": "https://github.com/pulumi/pulumi#readme",
  "_id": "@pulumi/pulumi@3.116.1",
  "_integrity": "sha512-Haweoa0O94MG+czV7+lcYJKunRCBxSz8VoHY23JWoCASWSc3i7/h4U8ejZB88GnrrRRQSihnKVCuz4Mb5Ngm+A==",
  "_resolved": "/home/runner/work/pulumi/pulumi/artifacts/sdk-nodejs-pulumi-pulumi-3.116.1.tgz",
  "_from": "file:../../artifacts/sdk-nodejs-pulumi-pulumi-3.116.1.tgz",
  "_nodeVersion": "16.20.2",
  "_npmVersion": "8.19.4",
  "dist": {
    "integrity": "sha512-Haweoa0O94MG+czV7+lcYJKunRCBxSz8VoHY23JWoCASWSc3i7/h4U8ejZB88GnrrRRQSihnKVCuz4Mb5Ngm+A==",
    "shasum": "226b662634635105c079fb6f583fa3eea92aefa8",
    "tarball": "https://registry.npmjs.org/@pulumi/pulumi/-/pulumi-3.116.1.tgz",
    "fileCount": 494,
    "unpackedSize": 14427261,
    "signatures": [
      {
        "keyid": "SHA256:jl3bwswu80PjjokCgh0o2w5c2U4LhQAE57gj9cz1kzA",
        "sig": "MEQCIHDRPDKsVOQV7TvjvCJ7IWETqK35d8tfyesh1X7SCKKJAiAFfjVGFCIxAeu1eeZoO+H9kM8Lc/MWWsEyDGStbX/Q6g=="
      }
    ]
  },
  "_npmUser": {
    "name": "pulumi-bot",
    "email": "bot@pulumi.com"
  },
  "directories": {

  },
  "maintainers": [
    {
      "name": "joeduffy",
      "email": "joe@pulumi.com"
    },
    {
      "name": "pulumi-bot",
      "email": "bot@pulumi.com"
    }
  ],
  "_npmOperationalInternal": {
    "host": "s3://npm-registry-packages",
    "tmp": "tmp/pulumi_3.116.1_1715761997764_0.7694246358483017"
  },
  "_hasShrinkwrap": false
}
				`,
			},
			want:    "https://github.com/pulumi/pulumi",
			wantErr: false,
		},
		{
			name: "fetchGitRepositoryFromNPM_error",

			args: args{
				packageName: "npm-package",
				result:      "",
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "fetchGitRepositoryFromNPM_error",

			args: args{
				packageName: "npm-package",
				result:      "foo",
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "fetchGitRepositoryFromNPM_error",

			args: args{
				packageName: "npm-package",
				result: `
{
  "objects": [],
  "total": 380,
  "time": "Wed Mar 09 2022 18:11:10 GMT+0000 (Coordinated Universal Time)"
}
				`,
			},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			p := pmc.NewMockClient(ctrl)
			p.EXPECT().Get(gomock.Any(), tt.args.packageName).
				DoAndReturn(func(url, packageName string) (*http.Response, error) {
					if tt.wantErr && tt.args.result == "" {
						return nil, errors.New("error")
					}

					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       io.NopCloser(bytes.NewBufferString(tt.args.result)),
					}, nil
				}).AnyTimes()
			got, err := fetchGitRepositoryFromNPM(tt.args.packageName, p)
			if (err != nil) != tt.wantErr {
				t.Errorf("fetchGitRepositoryFromNPM() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("fetchGitRepositoryFromNPM() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_findGitRepositoryInPYPIResponse(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                string
		partialPYPIResponse string
		want                string
		wantErrStr          string
	}{
		{
			name: "findGitRepositoryInPYPIResponse_none",
			partialPYPIResponse: `
 {
  "info": {
    "platform": "UNKNOWN",
    "not_a_project_url": "https://github.com/htaslan/color",
    "project_urls": {
      "Homepage": "http://git_NOT_VALID_hub.com/htaslan/color"
    }
  }
}
`,
			want:       "",
			wantErrStr: "could not find source repo for pypi package: somePackage",
		},
		{
			name: "findGitRepositoryInPYPIResponse_project_url",
			partialPYPIResponse: `
 {
  "info": {
    "platform": "UNKNOWN",
    "project_url": "https://github.com/htaslan/color/",
    "project_urls": {
      "Homepage": "http://git_NOT_VALID_hub.com/htaslan/color"
    }
  }
}
`,
			want:       "https://github.com/htaslan/color",
			wantErrStr: "",
		},
		{
			name: "findGitRepositoryInPYPIResponse_project_urls",
			partialPYPIResponse: `
 {
  "info": {
    "platform": "UNKNOWN",

    "project_url": "http://git_NOT_VALID_hub.com/htaslan/color",
    "project_urls": {
      "RandomKey": "https://github.com/htaslan/color/",
      "SponsorsIgnored": "https://github.com/sponsors/htaslan",
      "AnotherRandomKey": "http://git_NOT_VALID_hub.com/htaslan/color"
    }
  }
}
`,
			want:       "https://github.com/htaslan/color",
			wantErrStr: "",
		},
		{
			name: "findGitRepositoryInPYPIResponse_dedup",
			partialPYPIResponse: `
 {
  "info": {
    "platform": "UNKNOWN",
    "project_url": "foo",
    "project_urls": {
      "RandomKey": "https://github.com/htaslan/color/",
      "AnotherRandomKey": "http://htaslan.github.io/color",
      "CapsTestKey": "http://HTASLAN.github.io/cOLOr",
      "TrailingGit": "https://github.com/htaslan/color.git"
    }
  }
}
`,
			want:       "https://github.com/htaslan/color",
			wantErrStr: "",
		},
		{
			name: "findGitRepositoryInPYPIResponse_dedup_gitlab",
			partialPYPIResponse: `
 {
  "info": {
    "platform": "UNKNOWN",
    "project_url": "foo",
    "project_urls": {
      "RandomKey": "https://gitlab.com/htaslan/color/",
      "raNdoMkEY": "https://gitlab.com/hTASLan/color"
    }
  }
}
`,
			want:       "https://gitlab.com/htaslan/color",
			wantErrStr: "",
		},
		{
			name: "findGitRepositoryInPYPIResponse_toomany",
			partialPYPIResponse: `
 {
  "info": {
    "platform": "UNKNOWN",
    "project_url": "foo",
    "project_urls": {
      "RandomKey": "https://github.com/htaslan/color/",
      "AnotherRandomKey": "https://gitlab.com/htaslan/color"
    }
  }
}
`,
			want:       "",
			wantErrStr: "found too many possible source repos for pypi package: somePackage",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := findGitRepositoryInPYPIResponse("somePackage", strings.NewReader(tt.partialPYPIResponse))
			if err != nil && (!strings.Contains(err.Error(), tt.wantErrStr) || tt.wantErrStr == "") {
				t.Errorf("findGitRepositoryInPYPIResponse() error = \"%v\" did not contain "+
					"wantErrStr = \"%v\" testcase name %v",
					err, tt.wantErrStr, tt.name)
				return
			}
			if err == nil && tt.wantErrStr != "" {
				t.Errorf("findGitRepositoryInPYPIResponse() had nil error, but wanted "+
					"wantErrStr = \"%v\" testcase name %v",
					tt.wantErrStr, tt.name)
				return
			}

			if got != tt.want {
				t.Errorf("findGitRepositoryInPYPIResponse() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_fetchGitRepositoryFromPYPI(t *testing.T) {
	t.Parallel()
	type args struct {
		packageName string
		result      string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "fetchGitRepositoryFromPYPI",
			args: args{
				packageName: "some-package",
				result: `
{
  "info": {
    "author": "Hüseyin Tekinaslan",
    "author_email": "htaslan@bil.omu.edu.tr",
    "bugtrack_url": null,
    "classifiers": [
      "Development Status :: 5 - Production/Stable",
      "License :: OSI Approved :: MIT License",
      "Programming Language :: Python",
      "Programming Language :: Python :: 3",
      "Programming Language :: Python :: 3.2",
      "Programming Language :: Python :: 3.3",
      "Programming Language :: Python :: 3.4",
      "Programming Language :: Python :: 3.5",
      "Programming Language :: Python :: Implementation :: CPython",
      "Topic :: Software Development :: Libraries :: Python Modules"
    ],
    "description": "UNKNOWN",
    "description_content_type": null,
    "docs_url": null,
    "download_url": null,
    "downloads": {
      "last_day": -1,
      "last_month": -1,
      "last_week": -1
    },
    "home_page": "http://github.com/htaslan/color",
    "keywords": "colorize pycolorize color pycolor",
    "license": "MIT",
    "maintainer": null,
    "maintainer_email": null,
    "name": "color",
    "package_url": "https://pypi.org/project/color/",
    "platform": "UNKNOWN",
    "project_url": "https://pypi.org/project/color/",
    "project_urls": {
      "Homepage": "http://github.com/htaslan/color",
	  "Source": "foo"
    },
    "release_url": "https://pypi.org/project/color/0.1/",
    "requires_dist": null,
    "requires_python": null,
    "summary": "python module for colorize string",
    "version": "0.1",
    "yanked": false,
    "yanked_reason": null
  },
  "last_serial": 2041956,
  "releases": {
    "0.1": [
      {
        "comment_text": "a python module of colorize string",
        "digests": {
          "md5": "1a4577069c636b28d85052db9a384b95",
          "sha256": "de5b51fea834cb067631beaa1ec11d7753f1e3615e836e2e4c34dcf2b343eac2"
        },
        "downloads": -1,
        "filename": "color-0.1.1.tar.gz",
        "has_sig": false,
        "md5_digest": "1a4577069c636b28d85052db9a384b95",
        "packagetype": "sdist",
        "python_version": "source",
        "requires_python": null,
        "size": 3568,
        "upload_time": "2016-04-01T13:23:25",
        "upload_time_iso_8601": "2016-04-01T13:23:25.284973Z",
        "url": "https://files.pythonhosted.org/packages/88/04/0defd6f424e5bafb5abc75510cbe119a85d80b5505f1de5cd9a16d89ba8c/color-0.1.1.tar.gz",
        "yanked": false,
        "yanked_reason": null
      }
    ]
  },
  "urls": [
    {
      "comment_text": "a python module of colorize string",
      "digests": {
        "md5": "1a4577069c636b28d85052db9a384b95",
        "sha256": "de5b51fea834cb067631beaa1ec11d7753f1e3615e836e2e4c34dcf2b343eac2"
      },
      "downloads": -1,
      "filename": "color-0.1.1.tar.gz",
      "has_sig": false,
      "md5_digest": "1a4577069c636b28d85052db9a384b95",
      "packagetype": "sdist",
      "python_version": "source",
      "requires_python": null,
      "size": 3568,
      "upload_time": "2016-04-01T13:23:25",
      "upload_time_iso_8601": "2016-04-01T13:23:25.284973Z",
      "url": "https://files.pythonhosted.org/packages/88/04/0defd6f424e5bafb5abc75510cbe119a85d80b5505f1de5cd9a16d89ba8c/color-0.1.1.tar.gz",
      "yanked": false,
      "yanked_reason": null
    }
  ],
  "vulnerabilities": []
}

				`,
			},
			want:    "https://github.com/htaslan/color",
			wantErr: false,
		},
		{
			name: "fetchGitRepositoryFromPYPI_error",

			args: args{
				packageName: "pypi-package",
				result:      "",
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "fetchGitRepositoryFromPYPI_error",

			args: args{
				packageName: "pypi-package",
				result:      "foo",
			},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			p := pmc.NewMockClient(ctrl)
			p.EXPECT().Get(gomock.Any(), tt.args.packageName).
				DoAndReturn(func(url, packageName string) (*http.Response, error) {
					if tt.wantErr && tt.args.result == "" {
						return nil, errors.New("error")
					}

					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       io.NopCloser(bytes.NewBufferString(tt.args.result)),
					}, nil
				}).AnyTimes()
			got, err := fetchGitRepositoryFromPYPI(tt.args.packageName, p)
			if (err != nil) != tt.wantErr {
				t.Errorf("fetchGitRepositoryFromPYPI() error = %v, wantErr %v testcase name %v", err, tt.wantErr, tt.name)
				return
			}
			if got != tt.want {
				t.Errorf("fetchGitRepositoryFromPYPI() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_fetchGitRepositoryFromRubyGems(t *testing.T) {
	t.Parallel()
	type args struct {
		packageName string
		result      string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "fetchGitRepositoryFromPYPI",
			args: args{
				packageName: "npm-package",
				result: `
{
  "name": "color",
  "downloads": 8177801,
  "version": "1.8",
  "version_created_at": "2015-10-26T05:03:11.976Z",
  "version_downloads": 4558362,
  "platform": "ruby",
  "authors": "Austin Ziegler, Matt Lyon",
  "info": "Color is a Ruby library to provide basic RGB, CMYK, HSL, and other colourspace\nmanipulation support to applications that require it. It also provides 152\nnamed RGB colours (184 with spelling variations) that are commonly supported in\nHTML, SVG, and X11 applications. A technique for generating monochromatic\ncontrasting palettes is also included.\n\nThe Color library performs purely mathematical manipulation of the colours\nbased on colour theory without reference to colour profiles (such as sRGB or\nAdobe RGB). For most purposes, when working with RGB and HSL colour spaces,\nthis won't matter. Absolute colour spaces (like CIE L*a*b* and XYZ) and cannot\nbe reliably converted to relative colour spaces (like RGB) without colour\nprofiles.\n\nColor 1.8 adds an alpha parameter to all &lt;tt&gt;#css_rgba&lt;/tt&gt; calls, fixes a bug\nexposed by new constant lookup semantics in Ruby 2, and ensures that\n&lt;tt&gt;Color.equivalent?&lt;/tt&gt; can only be called on Color instances.\n\nBarring bugs introduced in this release, this (really) is the last version of\ncolor that supports Ruby 1.8, so make sure that your gem specification is set\nproperly (to &lt;tt&gt;~&gt; 1.8&lt;/tt&gt;) if that matters for your application. This\nversion will no longer be supported one year after the release of color 2.0.",
  "licenses": [
    "MIT"
  ],
  "metadata": {},
  "yanked": false,
  "sha": "0a8512ecf6a8fe14928707f7d2766680c955b3a2224de198c1e25c837cd36f82",
  "project_uri": "https://rubygems.org/gems/color",
  "gem_uri": "https://rubygems.org/gems/color-1.8.gem",
  "homepage_uri": "https://github.com/halostatue/color",
  "wiki_uri": null,
  "documentation_uri": "https://www.rubydoc.info/gems/color/1.8",
  "mailing_list_uri": null,
  "source_code_uri": "foo",
  "bug_tracker_uri": null,
  "changelog_uri": null,
  "funding_uri": null,
  "dependencies": {
    "development": [
      {
        "name": "hoe",
        "requirements": "~> 3.14"
      },
      {
        "name": "hoe-doofus",
        "requirements": "~> 1.0"
      },
      {
        "name": "hoe-gemspec2",
        "requirements": "~> 1.1"
      },
      {
        "name": "hoe-git",
        "requirements": "~> 1.6"
      },
      {
        "name": "hoe-travis",
        "requirements": "~> 1.2"
      },
      {
        "name": "minitest",
        "requirements": "~> 5.8"
      },
      {
        "name": "minitest-around",
        "requirements": "~> 0.3"
      },
      {
        "name": "minitest-autotest",
        "requirements": "~> 1.0"
      },
      {
        "name": "minitest-bisect",
        "requirements": "~> 1.2"
      },
      {
        "name": "minitest-focus",
        "requirements": "~> 1.1"
      },
      {
        "name": "minitest-moar",
        "requirements": "~> 0.0"
      },
      {
        "name": "minitest-pretty_diff",
        "requirements": "~> 0.1"
      },
      {
        "name": "rake",
        "requirements": "~> 10.0"
      },
      {
        "name": "rdoc",
        "requirements": "~> 4.0"
      },
      {
        "name": "simplecov",
        "requirements": "~> 0.7"
      }
    ],
    "runtime": []
  }
}
				`,
			},
			want:    "foo",
			wantErr: false,
		},
		{
			name: "fetchGitRepositoryFromNPM_error",

			args: args{
				packageName: "npm-package",
				result:      "",
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "fetchGitRepositoryFromNPM_error",

			args: args{
				packageName: "npm-package",
				result:      "foo",
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "empty project url",
			args: args{
				packageName: "npm-package",
				result: `
				{
  "name": "color",
  "downloads": 8177801,
  "version": "1.8",
  "version_created_at": "2015-10-26T05:03:11.976Z",
  "version_downloads": 4558362,
  "platform": "ruby",
  "authors": "Austin Ziegler, Matt Lyon",
  "info": "Color is a Ruby library to provide basic RGB, CMYK, HSL, and other colourspace\nmanipulation support to applications that require it. It also provides 152\nnamed RGB colours (184 with spelling variations) that are commonly supported in\nHTML, SVG, and X11 applications. A technique for generating monochromatic\ncontrasting palettes is also included.\n\nThe Color library performs purely mathematical manipulation of the colours\nbased on colour theory without reference to colour profiles (such as sRGB or\nAdobe RGB). For most purposes, when working with RGB and HSL colour spaces,\nthis won't matter. Absolute colour spaces (like CIE L*a*b* and XYZ) and cannot\nbe reliably converted to relative colour spaces (like RGB) without colour\nprofiles.\n\nColor 1.8 adds an alpha parameter to all &lt;tt&gt;#css_rgba&lt;/tt&gt; calls, fixes a bug\nexposed by new constant lookup semantics in Ruby 2, and ensures that\n&lt;tt&gt;Color.equivalent?&lt;/tt&gt; can only be called on Color instances.\n\nBarring bugs introduced in this release, this (really) is the last version of\ncolor that supports Ruby 1.8, so make sure that your gem specification is set\nproperly (to &lt;tt&gt;~&gt; 1.8&lt;/tt&gt;) if that matters for your application. This\nversion will no longer be supported one year after the release of color 2.0.",
  "licenses": [
    "MIT"
  ],
  "metadata": {},
  "yanked": false,
  "sha": "0a8512ecf6a8fe14928707f7d2766680c955b3a2224de198c1e25c837cd36f82",
  "project_uri": "https://rubygems.org/gems/color",
  "gem_uri": "https://rubygems.org/gems/color-1.8.gem",
  "homepage_uri": "https://github.com/halostatue/color",
  "wiki_uri": null,
  "documentation_uri": "https://www.rubydoc.info/gems/color/1.8",
  "mailing_list_uri": null,
  "source_code_uri": "",
  "bug_tracker_uri": null,
  "changelog_uri": null,
  "funding_uri": null,
  "dependencies": {
    "development": [
      {
        "name": "hoe",
        "requirements": "~> 3.14"
      },
      {
        "name": "hoe-doofus",
        "requirements": "~> 1.0"
      },
      {
        "name": "hoe-gemspec2",
        "requirements": "~> 1.1"
      },
      {
        "name": "hoe-git",
        "requirements": "~> 1.6"
      },
      {
        "name": "hoe-travis",
        "requirements": "~> 1.2"
      },
      {
        "name": "minitest",
        "requirements": "~> 5.8"
      },
      {
        "name": "minitest-around",
        "requirements": "~> 0.3"
      },
      {
        "name": "minitest-autotest",
        "requirements": "~> 1.0"
      },
      {
        "name": "minitest-bisect",
        "requirements": "~> 1.2"
      },
      {
        "name": "minitest-focus",
        "requirements": "~> 1.1"
      },
      {
        "name": "minitest-moar",
        "requirements": "~> 0.0"
      },
      {
        "name": "minitest-pretty_diff",
        "requirements": "~> 0.1"
      },
      {
        "name": "rake",
        "requirements": "~> 10.0"
      },
      {
        "name": "rdoc",
        "requirements": "~> 4.0"
      },
      {
        "name": "simplecov",
        "requirements": "~> 0.7"
      }
    ],
    "runtime": []
  }
}
				`,
			},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			p := pmc.NewMockClient(ctrl)
			p.EXPECT().Get(gomock.Any(), tt.args.packageName).
				DoAndReturn(func(url, packageName string) (*http.Response, error) {
					if tt.wantErr && tt.args.result == "" {
						return nil, errors.New("error")
					}

					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       io.NopCloser(bytes.NewBufferString(tt.args.result)),
					}, nil
				}).AnyTimes()
			got, err := fetchGitRepositoryFromRubyGems(tt.args.packageName, p)
			if (err != nil) != tt.wantErr {
				t.Errorf("fetchGitRepositoryFromRubyGems() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("fetchGitRepositoryFromRubyGems() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_fetchGitRepositoryFromNuget(t *testing.T) {
	t.Parallel()
	type args struct {
		packageName string
		result      string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "Return repository from nuget client",
			args: args{
				packageName: "nuget-package",
				result:      "nuget",
			},
			want:    "nuget",
			wantErr: false,
		},
		{
			name: "Error from nuget client",
			args: args{
				packageName: "nuget-package",
				result:      "",
			},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			n := ngt.NewMockClient(ctrl)
			n.EXPECT().GitRepositoryByPackageName(tt.args.packageName).
				DoAndReturn(func(packageName string) (string, error) {
					if tt.wantErr && tt.args.result == "" {
						return "", errors.New("error")
					}

					return tt.args.result, nil
				}).AnyTimes()
			got, err := fetchGitRepositoryFromNuget(tt.args.packageName, n)
			if (err != nil) != tt.wantErr {
				t.Errorf("fetchGitRepositoryFromNuget() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("fetchGitRepositoryFromNuget() = %v, want %v", got, tt.want)
			}
		})
	}
}
