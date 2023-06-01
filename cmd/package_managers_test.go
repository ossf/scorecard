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

// Package cmd implements Scorecard commandline.
package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"golang.org/x/exp/slices"
)

func Test_fetchGitRepositoryFromNPM(t *testing.T) {
	t.Parallel()
	type args struct {
		inputPackageName string
		result           string
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
				inputPackageName: "npm-package",
				result: `
{
  "objects": [
    {
      "package": {
        "name": "@pulumi/pulumi",
        "scope": "pulumi",
        "version": "3.26.0",
        "description": "Pulumi's Node.js SDK",
        "date": "2022-03-09T14:05:40.682Z",
        "links": {
          "homepage": "https://github.com/pulumi/pulumi#readme",
          "repository": "https://github.com/pulumi/pulumi",
          "bugs": "https://github.com/pulumi/pulumi/issues"
        },
        "publisher": {
          "username": "pulumi-bot",
          "email": "bot@pulumi.com"
        },
        "maintainers": [
          {
            "username": "joeduffy",
            "email": "joe@pulumi.com"
          },
          {
            "username": "pulumi-bot",
            "email": "bot@pulumi.com"
          }
        ]
      },
      "score": {
        "final": 0.4056031974977145,
        "detail": {
          "quality": 0.7308571951451065,
          "popularity": 0.19908392082147397,
          "maintenance": 0.3333333333333333
        }
      },
      "searchScore": 0.00090895034
    }
  ],
  "total": 380,
  "time": "Wed Mar 09 2022 18:11:10 GMT+0000 (Coordinated Universal Time)"
}
				`,
			},
			want:    "https://github.com/pulumi/pulumi",
			wantErr: false,
		},
		{
			name: "fetchGitRepositoryFromNPM_error",

			args: args{
				inputPackageName: "npm-package",
				result:           "",
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "fetchGitRepositoryFromNPM_error",

			args: args{
				inputPackageName: "npm-package",
				result:           "foo",
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "fetchGitRepositoryFromNPM_error",

			args: args{
				inputPackageName: "npm-package",
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
			p := NewMockpackageManagerClient(ctrl)
			p.EXPECT().Get(gomock.Any(), tt.args.inputPackageName).
				DoAndReturn(func(url, inputPackageName string) (*http.Response, error) {
					if tt.wantErr && tt.args.result == "" {
						//nolint
						return nil, errors.New("error")
					}

					return &http.Response{
						StatusCode: 200,
						Body:       io.NopCloser(bytes.NewBufferString(tt.args.result)),
					}, nil
				}).AnyTimes()
			got, err := fetchGitRepositoryFromNPM(tt.args.inputPackageName, p)
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

func Test_fetchGitRepositoryFromPYPI(t *testing.T) {
	t.Parallel()
	type args struct {
		inputPackageName string
		result           string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "fetchGitRepositoryFromPYPI",
			//nolint
			args: args{
				inputPackageName: "npm-package",
				//nolint
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
    "downoad_url": null,
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
			want:    "foo",
			wantErr: false,
		},
		{
			name: "fetchGitRepositoryFromNPM_error",

			args: args{
				inputPackageName: "npm-package",
				result:           "",
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "fetchGitRepositoryFromNPM_error",

			args: args{
				inputPackageName: "npm-package",
				result:           "foo",
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "empty project url",
			//nolint
			args: args{
				inputPackageName: "npm-package",
				//nolint
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
    "downoad_url": null,
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
	  "Source": ""
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
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			p := NewMockpackageManagerClient(ctrl)
			p.EXPECT().Get(gomock.Any(), tt.args.inputPackageName).
				DoAndReturn(func(url, inputPackageName string) (*http.Response, error) {
					if tt.wantErr && tt.args.result == "" {
						//nolint
						return nil, errors.New("error")
					}

					return &http.Response{
						StatusCode: 200,
						Body:       io.NopCloser(bytes.NewBufferString(tt.args.result)),
					}, nil
				}).AnyTimes()
			got, err := fetchGitRepositoryFromPYPI(tt.args.inputPackageName, p)
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
		inputPackageName string
		result           string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "fetchGitRepositoryFromPYPI",
			//nolint
			args: args{
				inputPackageName: "npm-package",
				//nolint
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
				inputPackageName: "npm-package",
				result:           "",
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "fetchGitRepositoryFromNPM_error",

			args: args{
				inputPackageName: "npm-package",
				result:           "foo",
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "empty project url",
			//nolint
			args: args{
				inputPackageName: "npm-package",
				//nolint
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
			p := NewMockpackageManagerClient(ctrl)
			p.EXPECT().Get(gomock.Any(), tt.args.inputPackageName).
				DoAndReturn(func(url, inputPackageName string) (*http.Response, error) {
					if tt.wantErr && tt.args.result == "" {
						//nolint
						return nil, errors.New("error")
					}

					return &http.Response{
						StatusCode: 200,
						Body:       io.NopCloser(bytes.NewBufferString(tt.args.result)),
					}, nil
				}).AnyTimes()
			got, err := fetchGitRepositoryFromRubyGems(tt.args.inputPackageName, p)
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

type resultPackagePage struct {
	url      string
	response string
}
type nugetTestArgs struct {
	inputPackageName               string
	expectedPackageName            string
	resultIndex                    string
	resultPackageRegistrationIndex string
	resultPackageSpec              string
	version                        string
	resultPackageRegistrationPages []resultPackagePage
}
type nugetTest struct {
	name    string
	want    string
	args    nugetTestArgs
	wantErr bool
}

func Test_fetchGitRepositoryFromNuget(t *testing.T) {
	t.Parallel()

	tests := []nugetTest{
		{
			name: "fetchGitRepositoryFromNuget_find_latest_version_in_single_page",

			args: nugetTestArgs{
				inputPackageName: "nuget-package",
				resultIndex: `
		    {
		      "version": "3.0.0",
		      "resources": [
		        {
		          "@id": "https://api.nuget.org/v3-flatcontainer/",
		          "@type": "PackageBaseAddress/3.0.0",
		          "comment": "Base URL of where NuGet packages are stored, in the format ..."
		        },
		        {
		          "@id": "https://api.nuget.org/v3/registration5-gz-semver1/",
		          "@type": "RegistrationsBaseUrl/3.6.0",
		          "comment": "Base URL of Azure storage where NuGet package registration info."
		        }
		        ]
		    }
		    `,
				resultPackageRegistrationIndex: `
		    {
		        "@id": "https://api.nuget.org/v3/registration5-semver1/Foo.NET/index.json",
		        "count": 1,
		        "items": [
		            {
		                "@id": "https://api.nuget.org/v3/registration5-semver1/Foo.NET/index.json#page/1",
		                "@type": "catalog:CatalogPage",
		                "count": 2,
		                "items": [
		                    {
		                        "@id": "https://api.nuget.org/v3/registration5-semver1/Foo.NET/3.5.8.json",
		                        "@type": "Package",
		                        "catalogEntry": {
		                            "@id": "https://api.nuget.org/v3/catalog0/data/2022.12.08.16.43.03/Foo.NET.3.5.8.json",
		                            "@type": "PackageDetails",
		                            "listed": true,
		                            "version": "3.5.8"
		                        }
		                    },
		                    {
		                        "@id": "https://api.nuget.org/v3/registration5-semver1/Foo.NET/4.0.1.json",
		                        "@type": "Package",
		                        "catalogEntry": {
		                            "@id": "https://api.nuget.org/v3/catalog0/data/2022.12.08.16.43.03/Foo.NET.4.0.1.json",
		                            "@type": "PackageDetails",
		                            "listed": true,
		                            "version": "4.0.1"
		                        }
		                    }
		                ]
		            }
		        ]
		    }
		    `,
				resultPackageRegistrationPages: []resultPackagePage{},
				resultPackageSpec: `
		    <package xmlns="http://schemas.microsoft.com/packaging/2013/05/nuspec.xsd">
		<metadata minClientVersion="2.12">
		      <id>Foo</id>
		      <version>4.0.1</version>
		      <title>Foo.NET</title>
		      <repository type="git" url="https://github.com/foo/foo.net" commit="123"/>
		  </metadata>
		      </package>
		    `,
				version: "4.0.1",
			},
			want:    "https://github.com/foo/foo.net",
			wantErr: false,
		},
		{
			name: "fetchGitRepositoryFromNuget_find_latest_version_with_lowercase",

			args: nugetTestArgs{
				inputPackageName:    "Nuget-Package",
				expectedPackageName: "nuget-package",
				resultIndex: `
		    {
		      "version": "3.0.0",
		      "resources": [
		        {
		          "@id": "https://api.nuget.org/v3-flatcontainer/",
		          "@type": "PackageBaseAddress/3.0.0",
		          "comment": "Base URL of where NuGet packages are stored, in the format ..."
		        },
		        {
		          "@id": "https://api.nuget.org/v3/registration5-gz-semver1/",
		          "@type": "RegistrationsBaseUrl/3.6.0",
		          "comment": "Base URL of Azure storage where NuGet package registration info."
		        }
		        ]
		    }
		    `,
				resultPackageRegistrationIndex: `
		    {
		        "@id": "https://api.nuget.org/v3/registration5-semver1/Foo.NET/index.json",
		        "count": 1,
		        "items": [
		            {
		                "@id": "https://api.nuget.org/v3/registration5-semver1/Foo.NET/index.json#page/1",
		                "@type": "catalog:CatalogPage",
		                "count": 2,
		                "items": [
		                    {
		                        "@id": "https://api.nuget.org/v3/registration5-semver1/Foo.NET/3.5.8.json",
		                        "@type": "Package",
		                        "catalogEntry": {
		                            "@id": "https://api.nuget.org/v3/catalog0/data/2022.12.08.16.43.03/Foo.NET.3.5.8.json",
		                            "@type": "PackageDetails",
		                            "listed": true,
		                            "version": "3.5.8"
		                        }
		                    },
		                    {
		                        "@id": "https://api.nuget.org/v3/registration5-semver1/Foo.NET/4.0.1.json",
		                        "@type": "Package",
		                        "catalogEntry": {
		                            "@id": "https://api.nuget.org/v3/catalog0/data/2022.12.08.16.43.03/Foo.NET.4.0.1.json",
		                            "@type": "PackageDetails",
		                            "listed": true,
		                            "version": "4.0.1"
		                        }
		                    }
		                ]
		            }
		        ]
		    }
		    `,
				resultPackageRegistrationPages: []resultPackagePage{},
				resultPackageSpec: `
		    <package xmlns="http://schemas.microsoft.com/packaging/2013/05/nuspec.xsd">
		<metadata minClientVersion="2.12">
		      <id>Foo</id>
		      <version>4.0.1</version>
		      <title>Foo.NET</title>
		      <repository type="git" url="https://github.com/foo/foo.net" commit="123"/>
		  </metadata>
		      </package>
		    `,
				version: "4.0.1",
			},
			want:    "https://github.com/foo/foo.net",
			wantErr: false,
		},
		{
			name: "fetchGitRepositoryFromNuget_find_latest_version_with_trailing_slash",

			args: nugetTestArgs{
				inputPackageName:    "Nuget-Package",
				expectedPackageName: "nuget-package",
				resultIndex: `
		    {
		      "version": "3.0.0",
		      "resources": [
		        {
		          "@id": "https://api.nuget.org/v3-flatcontainer/",
		          "@type": "PackageBaseAddress/3.0.0",
		          "comment": "Base URL of where NuGet packages are stored, in the format ..."
		        },
		        {
		          "@id": "https://api.nuget.org/v3/registration5-gz-semver1/",
		          "@type": "RegistrationsBaseUrl/3.6.0",
		          "comment": "Base URL of Azure storage where NuGet package registration info."
		        }
		        ]
		    }
		    `,
				resultPackageRegistrationIndex: `
		    {
		        "@id": "https://api.nuget.org/v3/registration5-semver1/Foo.NET/index.json",
		        "count": 1,
		        "items": [
		            {
		                "@id": "https://api.nuget.org/v3/registration5-semver1/Foo.NET/index.json#page/1",
		                "@type": "catalog:CatalogPage",
		                "count": 2,
		                "items": [
		                    {
		                        "@id": "https://api.nuget.org/v3/registration5-semver1/Foo.NET/3.5.8.json",
		                        "@type": "Package",
		                        "catalogEntry": {
		                            "@id": "https://api.nuget.org/v3/catalog0/data/2022.12.08.16.43.03/Foo.NET.3.5.8.json",
		                            "@type": "PackageDetails",
		                            "listed": true,
		                            "version": "3.5.8"
		                        }
		                    },
		                    {
		                        "@id": "https://api.nuget.org/v3/registration5-semver1/Foo.NET/4.0.1.json",
		                        "@type": "Package",
		                        "catalogEntry": {
		                            "@id": "https://api.nuget.org/v3/catalog0/data/2022.12.08.16.43.03/Foo.NET.4.0.1.json",
		                            "@type": "PackageDetails",
		                            "listed": true,
		                            "version": "4.0.1"
		                        }
		                    }
		                ]
		            }
		        ]
		    }
		    `,
				resultPackageRegistrationPages: []resultPackagePage{},
				resultPackageSpec: `
		    <package xmlns="http://schemas.microsoft.com/packaging/2013/05/nuspec.xsd">
		<metadata minClientVersion="2.12">
		      <id>Foo</id>
		      <version>4.0.1</version>
		      <title>Foo.NET</title>
		      <repository type="git" url="https://github.com/foo/foo.net/" commit="123"/>
		  </metadata>
		      </package>
		    `,
				version: "4.0.1",
			},
			want:    "https://github.com/foo/foo.net/",
			wantErr: false,
		},
		{
			name: "fetchGitRepositoryFromNuget_find_latest_version_without_git_ending",

			args: nugetTestArgs{
				inputPackageName: "nuget-package",
				resultIndex: `
		    {
		      "version": "3.0.0",
		      "resources": [
		        {
		          "@id": "https://api.nuget.org/v3-flatcontainer/",
		          "@type": "PackageBaseAddress/3.0.0",
		          "comment": "Base URL of where NuGet packages are stored, in the format ..."
		        },
		        {
		          "@id": "https://api.nuget.org/v3/registration5-gz-semver1/",
		          "@type": "RegistrationsBaseUrl/3.6.0",
		          "comment": "Base URL of Azure storage where NuGet package registration info."
		        }
		        ]
		    }
		    `,
				resultPackageRegistrationIndex: `
		    {
		        "@id": "https://api.nuget.org/v3/registration5-semver1/Foo.NET/index.json",
		        "count": 1,
		        "items": [
		            {
		                "@id": "https://api.nuget.org/v3/registration5-semver1/Foo.NET/index.json#page/1",
		                "@type": "catalog:CatalogPage",
		                "count": 2,
		                "items": [
		                    {
		                        "@id": "https://api.nuget.org/v3/registration5-semver1/Foo.NET/3.5.8.json",
		                        "@type": "Package",
		                        "catalogEntry": {
		                            "@id": "https://api.nuget.org/v3/catalog0/data/2022.12.08.16.43.03/Foo.NET.3.5.8.json",
		                            "@type": "PackageDetails",
		                            "listed": true,
		                            "version": "3.5.8"
		                        }
		                    },
		                    {
		                        "@id": "https://api.nuget.org/v3/registration5-semver1/Foo.NET/4.0.1.json",
		                        "@type": "Package",
		                        "catalogEntry": {
		                            "@id": "https://api.nuget.org/v3/catalog0/data/2022.12.08.16.43.03/Foo.NET.4.0.1.json",
		                            "@type": "PackageDetails",
		                            "listed": true,
		                            "version": "4.0.1"
		                        }
		                    }
		                ]
		            }
		        ]
		    }
		    `,
				resultPackageRegistrationPages: []resultPackagePage{},
				resultPackageSpec: `
		    <package xmlns="http://schemas.microsoft.com/packaging/2013/05/nuspec.xsd">
		<metadata minClientVersion="2.12">
		      <id>Foo</id>
		      <version>4.0.1</version>
		      <title>Foo.NET</title>
		      <repository type="git" url="https://github.com/foo/foo.net.git" commit="123"/>
		  </metadata>
		      </package>
		    `,
				version: "4.0.1",
			},
			want:    "https://github.com/foo/foo.net",
			wantErr: false,
		},
		{
			name: "fetchGitRepositoryFromNuget_find_latest_version_handle_four_digit_version",

			args: nugetTestArgs{
				inputPackageName: "nuget-package",
				resultIndex: `
		    {
		      "version": "3.0.0",
		      "resources": [
		        {
		          "@id": "https://api.nuget.org/v3-flatcontainer/",
		          "@type": "PackageBaseAddress/3.0.0",
		          "comment": "Base URL of where NuGet packages are stored, in the format ..."
		        },
		        {
		          "@id": "https://api.nuget.org/v3/registration5-gz-semver1/",
		          "@type": "RegistrationsBaseUrl/3.6.0",
		          "comment": "Base URL of Azure storage where NuGet package registration info."
		        }
		        ]
		    }
		    `,
				resultPackageRegistrationIndex: `
		    {
		        "@id": "https://api.nuget.org/v3/registration5-semver1/Foo.NET/index.json",
		        "count": 1,
		        "items": [
		            {
		                "@id": "https://api.nuget.org/v3/registration5-semver1/Foo.NET/index.json#page/1",
		                "@type": "catalog:CatalogPage",
		                "count": 2,
		                "items": [
		                    {
		                        "@id": "https://api.nuget.org/v3/registration5-semver1/Foo.NET/3.5.8.json",
		                        "@type": "Package",
		                        "catalogEntry": {
		                            "@id": "https://api.nuget.org/v3/catalog0/data/2022/Foo.NET.3.5.8.json",
		                            "@type": "PackageDetails",
		                            "listed": true,
		                            "version": "3.5.8"
		                        }
		                    },
		                    {
		                        "@id": "https://api.nuget.org/v3/registration5-semver1/Foo.NET/1.60.0.2981.json",
		                        "@type": "Package",
		                        "catalogEntry": {
		                            "@id": "https://api.nuget.org/v3/catalog0/data/2022/Foo.NET.1.60.0.2981.json",
		                            "@type": "PackageDetails",
		                            "listed": true,
		                            "version": "1.60.0.2981+metadata"
		                        }
		                    }
		                ]
		            }
		        ]
		    }
		    `,
				resultPackageRegistrationPages: []resultPackagePage{},
				resultPackageSpec: `
		    <package xmlns="http://schemas.microsoft.com/packaging/2013/05/nuspec.xsd">
		<metadata minClientVersion="2.12">
		      <id>Foo</id>
		      <version>1.60.0.2981+metadata</version>
		      <title>Foo.NET</title>
		      <repository type="git" url="https://github.com/foo/foo.net.git" commit="123"/>
		  </metadata>
		      </package>
		    `,
				version: "1.60.0.2981",
			},
			want:    "https://github.com/foo/foo.net",
			wantErr: false,
		},
		{
			name: "fetchGitRepositoryFromNuget_find_latest_version_skip_semver_metadataa",

			args: nugetTestArgs{
				inputPackageName: "nuget-package",
				resultIndex: `
		    {
		      "version": "3.0.0",
		      "resources": [
		        {
		          "@id": "https://api.nuget.org/v3-flatcontainer/",
		          "@type": "PackageBaseAddress/3.0.0",
		          "comment": "Base URL of where NuGet packages are stored, in the format ..."
		        },
		        {
		          "@id": "https://api.nuget.org/v3/registration5-gz-semver1/",
		          "@type": "RegistrationsBaseUrl/3.6.0",
		          "comment": "Base URL of Azure storage where NuGet package registration info."
		        }
		        ]
		    }
		    `,
				resultPackageRegistrationIndex: `
		    {
		        "@id": "https://api.nuget.org/v3/registration5-semver1/Foo.NET/index.json",
		        "count": 1,
		        "items": [
		            {
		                "@id": "https://api.nuget.org/v3/registration5-semver1/Foo.NET/index.json#page/1",
		                "@type": "catalog:CatalogPage",
		                "count": 2,
		                "items": [
		                    {
		                        "@id": "https://api.nuget.org/v3/registration5-semver1/Foo.NET/3.5.8.json",
		                        "@type": "Package",
		                        "catalogEntry": {
		                            "@id": "https://api.nuget.org/v3/catalog0/data/2022.12.08.16.43.03/Foo.NET.3.5.8.json",
		                            "@type": "PackageDetails",
		                            "listed": true,
		                            "version": "3.5.8"
		                        }
		                    },
		                    {
		                        "@id": "https://api.nuget.org/v3/registration5-semver1/Foo.NET/4.0.1.json",
		                        "@type": "Package",
		                        "catalogEntry": {
		                            "@id": "https://api.nuget.org/v3/catalog0/data/2022.12.08.16.43.03/Foo.NET.4.0.1.json",
		                            "@type": "PackageDetails",
		                            "listed": true,
		                            "version": "4.0.1+metadata"
		                        }
		                    }
		                ]
		            }
		        ]
		    }
		    `,
				resultPackageRegistrationPages: []resultPackagePage{},
				resultPackageSpec: `
		    <package xmlns="http://schemas.microsoft.com/packaging/2013/05/nuspec.xsd">
		<metadata minClientVersion="2.12">
		      <id>Foo</id>
		      <version>4.0.1</version>
		      <title>Foo.NET</title>
		      <repository type="git" url="https://github.com/foo/foo.net" commit="123"/>
		  </metadata>
		      </package>
		    `,
				version: "4.0.1",
			},
			want:    "https://github.com/foo/foo.net",
			wantErr: false,
		},
		{
			name: "fetchGitRepositoryFromNuget_find_latest_version_skip_pre_release",

			args: nugetTestArgs{
				inputPackageName: "nuget-package",
				resultIndex: `
		    {
		      "version": "3.0.0",
		      "resources": [
		        {
		          "@id": "https://api.nuget.org/v3-flatcontainer/",
		          "@type": "PackageBaseAddress/3.0.0",
		          "comment": "Base URL of where NuGet packages are stored, in the format ..."
		        },
		        {
		          "@id": "https://api.nuget.org/v3/registration5-gz-semver1/",
		          "@type": "RegistrationsBaseUrl/3.6.0",
		          "comment": "Base URL of Azure storage where NuGet package registration info."
		        }
		        ]
		    }
		    `,
				resultPackageRegistrationIndex: `
		    {
		        "@id": "https://api.nuget.org/v3/registration5-semver1/Foo.NET/index.json",
		        "count": 1,
		        "items": [
		            {
		                "@id": "https://api.nuget.org/v3/registration5-semver1/Foo.NET/index.json#page/1",
		                "@type": "catalog:CatalogPage",
		                "count": 2,
		                "items": [
		                    {
		                        "@id": "https://api.nuget.org/v3/registration5-semver1/Foo.NET/3.5.8.json",
		                        "@type": "Package",
		                        "catalogEntry": {
		                            "@id": "https://api.nuget.org/v3/catalog0/data/2022.12.08.16.43.03/Foo.NET.3.5.8.json",
		                            "@type": "PackageDetails",
		                            "listed": true,
		                            "version": "4.0.1"
		                        }
		                    },
		                    {
		                        "@id": "https://api.nuget.org/v3/registration5-semver1/Foo.NET/4.0.1.json",
		                        "@type": "Package",
		                        "catalogEntry": {
		                            "@id": "https://api.nuget.org/v3/catalog0/data/2022.12.08.16.43.03/Foo.NET.4.0.1.json",
		                            "@type": "PackageDetails",
		                            "listed": true,
		                            "version": "4.0.1-beta"
		                        }
		                    }
		                ]
		            }
		        ]
		    }
		    `,
				resultPackageRegistrationPages: []resultPackagePage{},
				resultPackageSpec: `
		    <package xmlns="http://schemas.microsoft.com/packaging/2013/05/nuspec.xsd">
		<metadata minClientVersion="2.12">
		      <id>Foo</id>
		      <version>4.0.1</version>
		      <title>Foo.NET</title>
		      <repository type="git" url="https://github.com/foo/foo.net" commit="123"/>
		  </metadata>
		      </package>
		    `,
				version: "4.0.1",
			},
			want:    "https://github.com/foo/foo.net",
			wantErr: false,
		},
		{
			name: "fetchGitRepositoryFromNuget_find_latest_version_skip_pre_release_and_metadata",

			args: nugetTestArgs{
				inputPackageName: "nuget-package",
				resultIndex: `
		    {
		      "version": "3.0.0",
		      "resources": [
		        {
		          "@id": "https://api.nuget.org/v3-flatcontainer/",
		          "@type": "PackageBaseAddress/3.0.0",
		          "comment": "Base URL of where NuGet packages are stored, in the format ..."
		        },
		        {
		          "@id": "https://api.nuget.org/v3/registration5-gz-semver1/",
		          "@type": "RegistrationsBaseUrl/3.6.0",
		          "comment": "Base URL of Azure storage where NuGet package registration info."
		        }
		        ]
		    }
		    `,
				resultPackageRegistrationIndex: `
		    {
		        "@id": "https://api.nuget.org/v3/registration5-semver1/Foo.NET/index.json",
		        "count": 1,
		        "items": [
		            {
		                "@id": "https://api.nuget.org/v3/registration5-semver1/Foo.NET/index.json#page/1",
		                "@type": "catalog:CatalogPage",
		                "count": 2,
		                "items": [
		                    {
		                        "@id": "https://api.nuget.org/v3/registration5-semver1/Foo.NET/3.5.8.json",
		                        "@type": "Package",
		                        "catalogEntry": {
		                            "@id": "https://api.nuget.org/v3/catalog0/data/2022.12.08.16.43.03/Foo.NET.3.5.8.json",
		                            "@type": "PackageDetails",
		                            "listed": true,
		                            "version": "4.0.1+metadata"
		                        }
		                    },
		                    {
		                        "@id": "https://api.nuget.org/v3/registration5-semver1/Foo.NET/4.0.1.json",
		                        "@type": "Package",
		                        "catalogEntry": {
		                            "@id": "https://api.nuget.org/v3/catalog0/data/2022.12.08.16.43.03/Foo.NET.4.0.1.json",
		                            "@type": "PackageDetails",
		                            "listed": true,
		                            "version": "4.0.1-beta+meta"
		                        }
		                    }
		                ]
		            }
		        ]
		    }
		    `,
				resultPackageRegistrationPages: []resultPackagePage{},
				resultPackageSpec: `
		    <package xmlns="http://schemas.microsoft.com/packaging/2013/05/nuspec.xsd">
		<metadata minClientVersion="2.12">
		      <id>Foo</id>
		      <version>4.0.1</version>
		      <title>Foo.NET</title>
		      <repository type="git" url="https://github.com/foo/foo.net" commit="123"/>
		  </metadata>
		      </package>
		    `,
				version: "4.0.1",
			},
			want:    "https://github.com/foo/foo.net",
			wantErr: false,
		},
		{
			name: "fetchGitRepositoryFromNuget_find_latest_version_get_github_project_url_if_repository_url_missing",

			args: nugetTestArgs{
				inputPackageName: "nuget-package",
				resultIndex: `
		    {
		      "version": "3.0.0",
		      "resources": [
		        {
		          "@id": "https://api.nuget.org/v3-flatcontainer/",
		          "@type": "PackageBaseAddress/3.0.0",
		          "comment": "Base URL of where NuGet packages are stored, in the format ..."
		        },
		        {
		          "@id": "https://api.nuget.org/v3/registration5-gz-semver1/",
		          "@type": "RegistrationsBaseUrl/3.6.0",
		          "comment": "Base URL of Azure storage where NuGet package registration info."
		        }
		        ]
		    }
		    `,
				resultPackageRegistrationIndex: `
		    {
		        "@id": "https://api.nuget.org/v3/registration5-semver1/Foo.NET/index.json",
		        "count": 1,
		        "items": [
		            {
		                "@id": "https://api.nuget.org/v3/registration5-semver1/Foo.NET/index.json#page/1",
		                "@type": "catalog:CatalogPage",
		                "count": 2,
		                "items": [
		                    {
		                        "@id": "https://api.nuget.org/v3/registration5-semver1/Foo.NET/3.5.8.json",
		                        "@type": "Package",
		                        "catalogEntry": {
		                            "@id": "https://api.nuget.org/v3/catalog0/data/2022.12.08.16.43.03/Foo.NET.3.5.8.json",
		                            "@type": "PackageDetails",
		                            "listed": true,
		                            "version": "3.5.8"
		                        }
		                    },
		                    {
		                        "@id": "https://api.nuget.org/v3/registration5-semver1/Foo.NET/4.0.1.json",
		                        "@type": "Package",
		                        "catalogEntry": {
		                            "@id": "https://api.nuget.org/v3/catalog0/data/2022.12.08.16.43.03/Foo.NET.4.0.1.json",
		                            "@type": "PackageDetails",
		                            "listed": true,
		                            "version": "4.0.1"
		                        }
		                    }
		                ]
		            }
		        ]
		    }
		    `,
				resultPackageRegistrationPages: []resultPackagePage{},
				resultPackageSpec: `
		    <package xmlns="http://schemas.microsoft.com/packaging/2013/05/nuspec.xsd">
		<metadata minClientVersion="2.12">
		      <id>Foo</id>
		      <version>4.0.1</version>
		      <title>Foo.NET</title>
		      <projectUrl>https://github.com/foo/foo.net</projectUrl>
		  </metadata>
		      </package>
		    `,
				version: "4.0.1",
			},
			want:    "https://github.com/foo/foo.net",
			wantErr: false,
		},
		{
			name: "fetchGitRepositoryFromNuget_find_latest_version_get_github_project_url_without_git_ending",

			args: nugetTestArgs{
				inputPackageName: "nuget-package",
				resultIndex: `
		    {
		      "version": "3.0.0",
		      "resources": [
		        {
		          "@id": "https://api.nuget.org/v3-flatcontainer/",
		          "@type": "PackageBaseAddress/3.0.0",
		          "comment": "Base URL of where NuGet packages are stored, in the format ..."
		        },
		        {
		          "@id": "https://api.nuget.org/v3/registration5-gz-semver1/",
		          "@type": "RegistrationsBaseUrl/3.6.0",
		          "comment": "Base URL of Azure storage where NuGet package registration info."
		        }
		        ]
		    }
		    `,
				resultPackageRegistrationIndex: `
		    {
		        "@id": "https://api.nuget.org/v3/registration5-semver1/Foo.NET/index.json",
		        "count": 1,
		        "items": [
		            {
		                "@id": "https://api.nuget.org/v3/registration5-semver1/Foo.NET/index.json#page/1",
		                "@type": "catalog:CatalogPage",
		                "count": 2,
		                "items": [
		                    {
		                        "@id": "https://api.nuget.org/v3/registration5-semver1/Foo.NET/3.5.8.json",
		                        "@type": "Package",
		                        "catalogEntry": {
		                            "@id": "https://api.nuget.org/v3/catalog0/data/2022.12.08.16.43.03/Foo.NET.3.5.8.json",
		                            "@type": "PackageDetails",
		                            "listed": true,
		                            "version": "3.5.8"
		                        }
		                    },
		                    {
		                        "@id": "https://api.nuget.org/v3/registration5-semver1/Foo.NET/4.0.1.json",
		                        "@type": "Package",
		                        "catalogEntry": {
		                            "@id": "https://api.nuget.org/v3/catalog0/data/2022.12.08.16.43.03/Foo.NET.4.0.1.json",
		                            "@type": "PackageDetails",
		                            "listed": true,
		                            "version": "4.0.1"
		                        }
		                    }
		                ]
		            }
		        ]
		    }
		    `,
				resultPackageRegistrationPages: []resultPackagePage{},
				resultPackageSpec: `
		    <package xmlns="http://schemas.microsoft.com/packaging/2013/05/nuspec.xsd">
		<metadata minClientVersion="2.12">
		      <id>Foo</id>
		      <version>4.0.1</version>
		      <title>Foo.NET</title>
		      <projectUrl>https://github.com/foo/foo.net.git</projectUrl>
		  </metadata>
		      </package>
		    `,
				version: "4.0.1",
			},
			want:    "https://github.com/foo/foo.net",
			wantErr: false,
		},
		{
			name: "fetchGitRepositoryFromNuget_find_latest_version_get_gitlab_project_url_if_repository_url_missing",

			args: nugetTestArgs{
				inputPackageName: "nuget-package",
				resultIndex: `
		    {
		      "version": "3.0.0",
		      "resources": [
		        {
		          "@id": "https://api.nuget.org/v3-flatcontainer/",
		          "@type": "PackageBaseAddress/3.0.0",
		          "comment": "Base URL of where NuGet packages are stored, in the format ..."
		        },
		        {
		          "@id": "https://api.nuget.org/v3/registration5-gz-semver1/",
		          "@type": "RegistrationsBaseUrl/3.6.0",
		          "comment": "Base URL of Azure storage where NuGet package registration info."
		        }
		        ]
		    }
		    `,
				resultPackageRegistrationIndex: `
		    {
		        "@id": "https://api.nuget.org/v3/registration5-semver1/Foo.NET/index.json",
		        "count": 1,
		        "items": [
		            {
		                "@id": "https://api.nuget.org/v3/registration5-semver1/Foo.NET/index.json#page/1",
		                "@type": "catalog:CatalogPage",
		                "count": 2,
		                "items": [
		                    {
		                        "@id": "https://api.nuget.org/v3/registration5-semver1/Foo.NET/3.5.8.json",
		                        "@type": "Package",
		                        "catalogEntry": {
		                            "@id": "https://api.nuget.org/v3/catalog0/data/2022.12.08.16.43.03/Foo.NET.3.5.8.json",
		                            "@type": "PackageDetails",
		                            "listed": true,
		                            "version": "3.5.8"
		                        }
		                    },
		                    {
		                        "@id": "https://api.nuget.org/v3/registration5-semver1/Foo.NET/4.0.1.json",
		                        "@type": "Package",
		                        "catalogEntry": {
		                            "@id": "https://api.nuget.org/v3/catalog0/data/2022.12.08.16.43.03/Foo.NET.4.0.1.json",
		                            "@type": "PackageDetails",
		                            "listed": true,
		                            "version": "4.0.1"
		                        }
		                    }
		                ]
		            }
		        ]
		    }
		    `,
				resultPackageRegistrationPages: []resultPackagePage{},
				resultPackageSpec: `
		    <package xmlns="http://schemas.microsoft.com/packaging/2013/05/nuspec.xsd">
		<metadata minClientVersion="2.12">
		      <id>Foo</id>
		      <version>4.0.1</version>
		      <title>Foo.NET</title>
		      <projectUrl>https://gitlab.com/foo/foo.net</projectUrl>
		  </metadata>
		      </package>
		    `,
				version: "4.0.1",
			},
			want:    "https://gitlab.com/foo/foo.net",
			wantErr: false,
		},
		{
			name: "fetchGitRepositoryFromNuget_find_latest_version_error_if_project_url_is_not_gitXXb",

			args: nugetTestArgs{
				inputPackageName: "nuget-package",
				resultIndex: `
		    {
		      "version": "3.0.0",
		      "resources": [
		        {
		          "@id": "https://api.nuget.org/v3-flatcontainer/",
		          "@type": "PackageBaseAddress/3.0.0",
		          "comment": "Base URL of where NuGet packages are stored, in the format ..."
		        },
		        {
		          "@id": "https://api.nuget.org/v3/registration5-gz-semver1/",
		          "@type": "RegistrationsBaseUrl/3.6.0",
		          "comment": "Base URL of Azure storage where NuGet package registration info."
		        }
		        ]
		    }
		    `,
				resultPackageRegistrationIndex: `
		    {
		        "@id": "https://api.nuget.org/v3/registration5-semver1/Foo.NET/index.json",
		        "count": 1,
		        "items": [
		            {
		                "@id": "https://api.nuget.org/v3/registration5-semver1/Foo.NET/index.json#page/1",
		                "@type": "catalog:CatalogPage",
		                "count": 2,
		                "items": [
		                    {
		                        "@id": "https://api.nuget.org/v3/registration5-semver1/Foo.NET/3.5.8.json",
		                        "@type": "Package",
		                        "catalogEntry": {
		                            "@id": "https://api.nuget.org/v3/catalog0/data/2022.12.08.16.43.03/Foo.NET.3.5.8.json",
		                            "@type": "PackageDetails",
		                            "listed": true,
		                            "version": "3.5.8"
		                        }
		                    },
		                    {
		                        "@id": "https://api.nuget.org/v3/registration5-semver1/Foo.NET/4.0.1.json",
		                        "@type": "Package",
		                        "catalogEntry": {
		                            "@id": "https://api.nuget.org/v3/catalog0/data/2022.12.08.16.43.03/Foo.NET.4.0.1.json",
		                            "@type": "PackageDetails",
		                            "listed": true,
		                            "version": "4.0.1"
		                        }
		                    }
		                ]
		            }
		        ]
		    }
		    `,
				resultPackageRegistrationPages: []resultPackagePage{},
				resultPackageSpec: `
		    <package xmlns="http://schemas.microsoft.com/packaging/2013/05/nuspec.xsd">
		<metadata minClientVersion="2.12">
		      <id>Foo</id>
		      <version>4.0.1</version>
		      <title>Foo.NET</title>
		      <projectUrl>https://myserver.com/foo/foo.net</projectUrl>
		  </metadata>
		      </package>
		    `,
				version: "4.0.1",
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "fetchGitRepositoryFromNuget_find_latest_version_in_first_of_multiple_pages",

			args: nugetTestArgs{
				inputPackageName: "nuget-package",
				resultIndex: `
		    {
		      "version": "3.0.0",
		      "resources": [
		        {
		          "@id": "https://api.nuget.org/v3-flatcontainer/",
		          "@type": "PackageBaseAddress/3.0.0",
		          "comment": "Base URL of where NuGet packages are stored, in the format ..."
		        },
		        {
		          "@id": "https://api.nuget.org/v3/registration5-gz-semver1/",
		          "@type": "RegistrationsBaseUrl/3.6.0",
		          "comment": "Base URL of Azure storage where NuGet package registration info."
		        }
		        ]
		    }
		    `,
				resultPackageRegistrationIndex: `
		    {
		        "@id": "https://api.nuget.org/v3/registration5-semver1/Foo.NET/index.json",
		        "count": 2,
		        "items": [
		            {
		                "@id": "https://api.nuget.org/v3/registration5-semver1/Foo.NET/index.json#page/1",
		                "@type": "catalog:CatalogPage",
		                "count": 2,
		                "items": [
		                    {
		                        "@id": "https://api.nuget.org/v3/registration5-semver1/Foo.NET/3.5.1.json",
		                        "@type": "Package",
		                        "catalogEntry": {
		                            "@id": "https://api.nuget.org/v3/catalog0/data/2022.12.08.16.43.03/Foo.NET.3.5.1.json",
		                            "@type": "PackageDetails",
		                            "listed": true,
		                            "version": "3.5.1"
		                        }
		                    },
		                    {
		                        "@id": "https://api.nuget.org/v3/registration5-semver1/Foo.NET/3.5.2.json",
		                        "@type": "Package",
		                        "catalogEntry": {
		                            "@id": "https://api.nuget.org/v3/catalog0/data/2022.12.08.16.43.03/Foo.NET.3.5.2.json",
		                            "@type": "PackageDetails",
		                            "listed": true,
		                            "version": "3.5.2"
		                        }
		                    }
		                ]
		            },
				{
		                "@id": "https://api.nuget.org/v3/registration5-semver1/Foo.NET/index.json#page/2",
		                "@type": "catalog:CatalogPage",
		                "count": 2,
		                "items": [
		                    {
		                        "@id": "https://api.nuget.org/v3/registration5-semver1/Foo.NET/3.5.8.json",
		                        "@type": "Package",
		                        "catalogEntry": {
		                            "@id": "https://api.nuget.org/v3/catalog0/data/2022.12.08.16.43.03/Foo.NET.3.5.8.json",
		                            "@type": "PackageDetails",
		                            "listed": true,
		                            "version": "3.5.8"
		                        }
		                    },
		                    {
		                        "@id": "https://api.nuget.org/v3/registration5-semver1/Foo.NET/4.0.1.json",
		                        "@type": "Package",
		                        "catalogEntry": {
		                            "@id": "https://api.nuget.org/v3/catalog0/data/2022.12.08.16.43.03/Foo.NET.4.0.1.json",
		                            "@type": "PackageDetails",
		                            "listed": true,
		                            "version": "4.0.1"
		                        }
		                    }
		                ]
		            }
		        ]
		    }
		    `,
				resultPackageRegistrationPages: []resultPackagePage{},
				resultPackageSpec: `
		    <package xmlns="http://schemas.microsoft.com/packaging/2013/05/nuspec.xsd">
		<metadata minClientVersion="2.12">
		      <id>Foo</id>
		      <version>4.0.1</version>
		      <title>Foo.NET</title>
		      <repository type="git" url="https://github.com/foo/foo.net" commit="123"/>
		  </metadata>
		      </package>
		    `,
				version: "4.0.1",
			},
			want:    "https://github.com/foo/foo.net",
			wantErr: false,
		},
		{
			name: "fetchGitRepositoryFromNuget_find_latest_version_in_first_of_multiple_remote_pages",

			args: nugetTestArgs{
				inputPackageName: "nuget-package",
				resultIndex: `
		    {
		      "version": "3.0.0",
		      "resources": [
		        {
		          "@id": "https://api.nuget.org/v3-flatcontainer/",
		          "@type": "PackageBaseAddress/3.0.0",
		          "comment": "Base URL of where NuGet packages are stored, in the format ..."
		        },
		        {
		          "@id": "https://api.nuget.org/v3/registration5-gz-semver1/",
		          "@type": "RegistrationsBaseUrl/3.6.0",
		          "comment": "Base URL of Azure storage where NuGet package registration info."
		        }
		        ]
		    }
		    `,
				resultPackageRegistrationIndex: `
		    {
		        "@id": "https://api.nuget.org/v3/registration5-semver1/Foo.NET/index.json",
		        "count": 2,
		        "items": [
		            {
		                "@id": "https://api.nuget.org/v3/registration5-semver1/Foo.NET/page1/index.json",
		                "@type": "catalog:CatalogPage",
		                "count": 2
		            },
				{
		                "@id": "https://api.nuget.org/v3/registration5-semver1/Foo.NET/page2/index.json",
		                "@type": "catalog:CatalogPage",
		                "count": 2
		            }
		        ]
		    }
		    `,
				resultPackageRegistrationPages: []resultPackagePage{
					{
						url: "https://api.nuget.org/v3/registration5-semver1/Foo.NET/page1/index.json",
						response: `
						{
							"@id": "https://api.nuget.org/v3/registration5-semver1/Foo.NET/index.json#page/1",
							"@type": "catalog:CatalogPage",
							"count": 2,
							"items": [
								{
									"@id": "https://api.nuget.org/v3/registration5-semver1/Foo.NET/3.5.1.json",
									"@type": "Package",
									"catalogEntry": {
										"@id": "https://api.nuget.org/v3/catalog0/data/2022.12.08.16.43.03/Foo.NET.3.5.1.json",
										"@type": "PackageDetails",
										"listed": true,
										"version": "3.5.1"
									}
								},
								{
									"@id": "https://api.nuget.org/v3/registration5-semver1/Foo.NET/3.5.2.json",
									"@type": "Package",
									"catalogEntry": {
										"@id": "https://api.nuget.org/v3/catalog0/data/2022.12.08.16.43.03/Foo.NET.3.5.2.json",
										"@type": "PackageDetails",
										"listed": true,
										"version": "3.5.2"
									}
								}
							]
						}
					`,
					},
					{
						url: "https://api.nuget.org/v3/registration5-semver1/Foo.NET/page2/index.json",
						response: `
						{
							"@id": "https://api.nuget.org/v3/registration5-semver1/Foo.NET/index.json#page/2",
							"@type": "catalog:CatalogPage",
							"count": 2,
							"items": [
								{
									"@id": "https://api.nuget.org/v3/registration5-semver1/Foo.NET/3.5.8.json",
									"@type": "Package",
									"catalogEntry": {
										"@id": "https://api.nuget.org/v3/catalog0/data/2022.12.08.16.43.03/Foo.NET.3.5.8.json",
										"@type": "PackageDetails",
										"listed": true,
										"version": "3.5.8"
									}
								},
								{
									"@id": "https://api.nuget.org/v3/registration5-semver1/Foo.NET/4.0.1.json",
									"@type": "Package",
									"catalogEntry": {
										"@id": "https://api.nuget.org/v3/catalog0/data/2022.12.08.16.43.03/Foo.NET.4.0.1.json",
										"@type": "PackageDetails",
										"listed": true,
										"version": "4.0.1"
									}
								}
							]
						}
						`,
					},
				},
				resultPackageSpec: `
		    <package xmlns="http://schemas.microsoft.com/packaging/2013/05/nuspec.xsd">
		<metadata minClientVersion="2.12">
		      <id>Foo</id>
		      <version>4.0.1</version>
		      <title>Foo.NET</title>
		      <repository type="git" url="https://github.com/foo/foo.net" commit="123"/>
		  </metadata>
		      </package>
		    `,
				version: "4.0.1",
			},
			want:    "https://github.com/foo/foo.net",
			wantErr: false,
		},
		{
			name: "fetchGitRepositoryFromNuget_find_latest_version_in_last_of_multiple_pages",

			args: nugetTestArgs{
				inputPackageName: "nuget-package",
				resultIndex: `
		    {
		      "version": "3.0.0",
		      "resources": [
		        {
		          "@id": "https://api.nuget.org/v3-flatcontainer/",
		          "@type": "PackageBaseAddress/3.0.0",
		          "comment": "Base URL of where NuGet packages are stored, in the format ..."
		        },
		        {
		          "@id": "https://api.nuget.org/v3/registration5-gz-semver1/",
		          "@type": "RegistrationsBaseUrl/3.6.0",
		          "comment": "Base URL of Azure storage where NuGet package registration info."
		        }
		        ]
		    }
		    `,
				resultPackageRegistrationIndex: `
		    {
		        "@id": "https://api.nuget.org/v3/registration5-semver1/Foo.NET/index.json",
		        "count": 2,
		        "items": [
		            {
		                "@id": "https://api.nuget.org/v3/registration5-semver1/Foo.NET/index.json#page/1",
		                "@type": "catalog:CatalogPage",
		                "count": 2,
		                "items": [
		                    {
		                        "@id": "https://api.nuget.org/v3/registration5-semver1/Foo.NET/3.5.8.json",
		                        "@type": "Package",
		                        "catalogEntry": {
		                            "@id": "https://api.nuget.org/v3/catalog0/data/2022.12.08.16.43.03/Foo.NET.3.5.8.json",
		                            "@type": "PackageDetails",
		                            "listed": true,
		                            "version": "3.5.8"
		                        }
		                    },
		                    {
		                        "@id": "https://api.nuget.org/v3/registration5-semver1/Foo.NET/4.0.1.json",
		                        "@type": "Package",
		                        "catalogEntry": {
		                            "@id": "https://api.nuget.org/v3/catalog0/data/2022.12.08.16.43.03/Foo.NET.4.0.1.json",
		                            "@type": "PackageDetails",
		                            "listed": true,
		                            "version": "4.0.1"
		                        }
		                    }
		                ]
		            },
				{
		                "@id": "https://api.nuget.org/v3/registration5-semver1/Foo.NET/index.json#page/2",
		                "@type": "catalog:CatalogPage",
		                "count": 2,
		                "items": [
		                    {
		                        "@id": "https://api.nuget.org/v3/registration5-semver1/Foo.NET/4.1.json",
		                        "@type": "Package",
		                        "catalogEntry": {
		                            "@id": "https://api.nuget.org/v3/catalog0/data/2022.12.08.16.43.03/Foo.NET.4.1.json",
		                            "@type": "PackageDetails",
		                            "listed": false,
		                            "version": "4.1"
		                        }
		                    },
		                    {
		                        "@id": "https://api.nuget.org/v3/registration5-semver1/Foo.NET/4.2.json",
		                        "@type": "Package",
		                        "catalogEntry": {
		                            "@id": "https://api.nuget.org/v3/catalog0/data/2022.12.08.16.43.03/Foo.NET.4.2.json",
		                            "@type": "PackageDetails",
		                            "listed": false,
		                            "version": "4.2"
		                        }
		                    }
		                ]
		            }
		        ]
		    }
		    `,
				resultPackageRegistrationPages: []resultPackagePage{},
				resultPackageSpec: `
		    <package xmlns="http://schemas.microsoft.com/packaging/2013/05/nuspec.xsd">
		<metadata minClientVersion="2.12">
		      <id>Foo</id>
		      <version>4.0.1</version>
		      <title>Foo.NET</title>
		      <repository type="git" url="https://github.com/foo/foo.net" commit="123"/>
		  </metadata>
		      </package>
		    `,
				version: "4.0.1",
			},
			want:    "https://github.com/foo/foo.net",
			wantErr: false,
		},
		{
			name: "fetchGitRepositoryFromNuget_find_latest_version_in_last_of_remote_multiple_pages",

			args: nugetTestArgs{
				inputPackageName: "nuget-package",
				resultIndex: `
		    {
		      "version": "3.0.0",
		      "resources": [
		        {
		          "@id": "https://api.nuget.org/v3-flatcontainer/",
		          "@type": "PackageBaseAddress/3.0.0",
		          "comment": "Base URL of where NuGet packages are stored, in the format ..."
		        },
		        {
		          "@id": "https://api.nuget.org/v3/registration5-gz-semver1/",
		          "@type": "RegistrationsBaseUrl/3.6.0",
		          "comment": "Base URL of Azure storage where NuGet package registration info."
		        }
		        ]
		    }
		    `,
				resultPackageRegistrationIndex: `
		    {
		        "@id": "https://api.nuget.org/v3/registration5-semver1/Foo.NET/index.json",
		        "count": 2,
		        "items": [
		            {
		                "@id": "https://api.nuget.org/v3/registration5-semver1/Foo.NET/page1/index.json",
		                "@type": "catalog:CatalogPage",
		                "count": 2
		            },
				{
		                "@id": "https://api.nuget.org/v3/registration5-semver1/Foo.NET/page2/index.json",
		                "@type": "catalog:CatalogPage",
		                "count": 2
		            }
		        ]
		    }
		    `,
				resultPackageRegistrationPages: []resultPackagePage{
					{
						url: "https://api.nuget.org/v3/registration5-semver1/Foo.NET/page1/index.json",
						response: `
						{
							"@id": "https://api.nuget.org/v3/registration5-semver1/Foo.NET/index.json#page/1",
							"@type": "catalog:CatalogPage",
							"count": 2,
							"items": [
								{
									"@id": "https://api.nuget.org/v3/registration5-semver1/Foo.NET/3.5.8.json",
									"@type": "Package",
									"catalogEntry": {
										"@id": "https://api.nuget.org/v3/catalog0/data/2022.12.08.16.43.03/Foo.NET.3.5.8.json",
										"@type": "PackageDetails",
										"listed": true,
										"version": "3.5.8"
									}
								},
								{
									"@id": "https://api.nuget.org/v3/registration5-semver1/Foo.NET/4.0.1.json",
									"@type": "Package",
									"catalogEntry": {
										"@id": "https://api.nuget.org/v3/catalog0/data/2022.12.08.16.43.03/Foo.NET.4.0.1.json",
										"@type": "PackageDetails",
										"listed": true,
										"version": "4.0.1"
									}
								}
							]
						}
					`,
					},
					{
						url: "https://api.nuget.org/v3/registration5-semver1/Foo.NET/page2/index.json",
						response: `
						{
							"@id": "https://api.nuget.org/v3/registration5-semver1/Foo.NET/index.json#page/2",
							"@type": "catalog:CatalogPage",
							"count": 2,
							"items": [
								{
									"@id": "https://api.nuget.org/v3/registration5-semver1/Foo.NET/4.1.json",
									"@type": "Package",
									"catalogEntry": {
										"@id": "https://api.nuget.org/v3/catalog0/data/2022.12.08.16.43.03/Foo.NET.4.1.json",
										"@type": "PackageDetails",
										"listed": false,
										"version": "4.1"
									}
								},
								{
									"@id": "https://api.nuget.org/v3/registration5-semver1/Foo.NET/4.2.json",
									"@type": "Package",
									"catalogEntry": {
										"@id": "https://api.nuget.org/v3/catalog0/data/2022.12.08.16.43.03/Foo.NET.4.2.json",
										"@type": "PackageDetails",
										"listed": false,
										"version": "4.2"
									}
								}
							]
						}
						`,
					},
				},
				resultPackageSpec: `
		    <package xmlns="http://schemas.microsoft.com/packaging/2013/05/nuspec.xsd">
		<metadata minClientVersion="2.12">
		      <id>Foo</id>
		      <version>4.0.1</version>
		      <title>Foo.NET</title>
		      <repository type="git" url="https://github.com/foo/foo.net" commit="123"/>
		  </metadata>
		      </package>
		    `,
				version: "4.0.1",
			},
			want:    "https://github.com/foo/foo.net",
			wantErr: false,
		},
		{
			name: "fetchGitRepositoryFromNuget_find_latest_version_with_default_listed_value_true",

			args: nugetTestArgs{
				inputPackageName: "nuget-package",
				resultIndex: `
		    {
		      "version": "3.0.0",
		      "resources": [
		        {
		          "@id": "https://api.nuget.org/v3-flatcontainer/",
		          "@type": "PackageBaseAddress/3.0.0",
		          "comment": "Base URL of where NuGet packages are stored, in the format ..."
		        },
		        {
		          "@id": "https://api.nuget.org/v3/registration5-gz-semver1/",
		          "@type": "RegistrationsBaseUrl/3.6.0",
		          "comment": "Base URL of Azure storage where NuGet package registration info is stored in GZIP format."
		        }
		        ]
		    }
		    `,
				resultPackageRegistrationIndex: `
		    {
		        "@id": "https://api.nuget.org/v3/registration5-semver1/Foo.NET/index.json",
		        "count": 1,
		        "items": [
		            {
		                "@id": "https://api.nuget.org/v3/registration5-semver1/Foo.NET/index.json#page/1",
		                "@type": "catalog:CatalogPage",
		                "count": 2,
		                "items": [
		                    {
		                        "@id": "https://api.nuget.org/v3/registration5-semver1/Foo.NET/3.5.8.json",
		                        "@type": "Package",
		                        "catalogEntry": {
		                            "@id": "https://api.nuget.org/v3/catalog0/data/2022.12.08.16.43.03/Foo.NET.3.5.8.json",
		                            "@type": "PackageDetails",
		                            "listed": true,
		                            "version": "3.5.8"
		                        }
		                    },
		                    {
		                        "@id": "https://api.nuget.org/v3/registration5-semver1/Foo.NET/4.0.1.json",
		                        "@type": "Package",
		                        "catalogEntry": {
		                            "@id": "https://api.nuget.org/v3/catalog0/data/2022.12.08.16.43.03/Foo.NET.4.0.1.json",
		                            "@type": "PackageDetails",
		                            "version": "4.0.1"
		                        }
		                    }
		                ]
		            }
		        ]
		    }
		    `,
				resultPackageRegistrationPages: []resultPackagePage{},
				resultPackageSpec: `
		    <package xmlns="http://schemas.microsoft.com/packaging/2013/05/nuspec.xsd">
		<metadata minClientVersion="2.12">
		      <id>Foo</id>
		      <version>4.0.1</version>
		      <title>Foo.NET</title>
		      <repository type="git" url="https://github.com/foo/foo.net" commit="123"/>
		  </metadata>
		      </package>
		    `,
				version: "4.0.1",
			},
			want:    "https://github.com/foo/foo.net",
			wantErr: false,
		},
		{
			name: "fetchGitRepositoryFromNuget_find_latest_version_with_skip_not_listed",

			args: nugetTestArgs{
				inputPackageName: "nuget-package",
				resultIndex: `
		    {
		      "version": "3.0.0",
		      "resources": [
		        {
		          "@id": "https://api.nuget.org/v3-flatcontainer/",
		          "@type": "PackageBaseAddress/3.0.0",
		          "comment": "Base URL of where NuGet packages are stored, in the format ..."
		        },
		        {
		          "@id": "https://api.nuget.org/v3/registration5-gz-semver1/",
		          "@type": "RegistrationsBaseUrl/3.6.0",
		          "comment": "Base URL of Azure storage where NuGet package registration info is stored in GZIP format."
		        }
		        ]
		    }
		    `,
				resultPackageRegistrationIndex: `
		    {
		        "@id": "https://api.nuget.org/v3/registration5-semver1/Foo.NET/index.json",
		        "count": 1,
		        "items": [
		            {
		                "@id": "https://api.nuget.org/v3/registration5-semver1/Foo.NET/index.json#page/1",
		                "@type": "catalog:CatalogPage",
		                "count": 2,
		                "items": [
		                    {
		                        "@id": "https://api.nuget.org/v3/registration5-semver1/Foo.NET/3.5.8.json",
		                        "@type": "Package",
		                        "catalogEntry": {
		                            "@id": "https://api.nuget.org/v3/catalog0/data/2022.12.08.16.43.03/Foo.NET.3.5.8.json",
		                            "@type": "PackageDetails",
		                            "listed": true,
		                            "version": "3.5.8"
		                        }
		                    },
		                    {
		                        "@id": "https://api.nuget.org/v3/registration5-semver1/Foo.NET/4.0.1.json",
		                        "@type": "Package",
		                        "catalogEntry": {
		                            "@id": "https://api.nuget.org/v3/catalog0/data/2022.12.08.16.43.03/Foo.NET.4.0.1.json",
		                            "@type": "PackageDetails",
		                            "listed": false,
		                            "version": "4.0.1"
		                        }
		                    }
		                ]
		            }
		        ]
		    }
		    `,
				resultPackageRegistrationPages: []resultPackagePage{},
				resultPackageSpec: `
		    <package xmlns="http://schemas.microsoft.com/packaging/2013/05/nuspec.xsd">
		<metadata minClientVersion="2.12">
		      <id>Foo</id>
		      <version>3.5.8</version>
		      <title>Foo.NET</title>
		      <repository type="git" url="https://github.com/foo/foo.net" commit="123"/>
		  </metadata>
		      </package>
		    `,
				version: "3.5.8",
			},
			want:    "https://github.com/foo/foo.net",
			wantErr: false,
		},
		{
			name: "fetchGitRepositoryFromNuget_no_listed_version",

			args: nugetTestArgs{
				inputPackageName: "nuget-package",
				resultIndex: `
		    {
		      "version": "3.0.0",
		      "resources": [
		        {
		          "@id": "https://api.nuget.org/v3-flatcontainer/",
		          "@type": "PackageBaseAddress/3.0.0",
		          "comment": "Base URL of where NuGet packages are stored, in the format ..."
		        },
		        {
		          "@id": "https://api.nuget.org/v3/registration5-gz-semver1/",
		          "@type": "RegistrationsBaseUrl/3.6.0",
		          "comment": "Base URL of Azure storage where NuGet package registration info is stored in GZIP format."
		        }
		        ]
		    }
		    `,
				resultPackageRegistrationIndex: `
		    {
		        "@id": "https://api.nuget.org/v3/registration5-semver1/Foo.NET/index.json",
		        "count": 1,
		        "items": [
		            {
		                "@id": "https://api.nuget.org/v3/registration5-semver1/Foo.NET/index.json#page/1",
		                "@type": "catalog:CatalogPage",
		                "count": 2,
		                "items": [
		                    {
		                        "@id": "https://api.nuget.org/v3/registration5-semver1/Foo.NET/3.5.8.json",
		                        "@type": "Package",
		                        "catalogEntry": {
		                            "@id": "https://api.nuget.org/v3/catalog0/data/2022.12.08.16.43.03/Foo.NET.3.5.8.json",
		                            "@type": "PackageDetails",
		                            "listed": false,
		                            "version": "3.5.8"
		                        }
		                    },
		                    {
		                        "@id": "https://api.nuget.org/v3/registration5-semver1/Foo.NET/4.0.1.json",
		                        "@type": "Package",
		                        "catalogEntry": {
		                            "@id": "https://api.nuget.org/v3/catalog0/data/2022.12.08.16.43.03/Foo.NET.4.0.1.json",
		                            "@type": "PackageDetails",
		                            "listed": false,
		                            "version": "4.0.1"
		                        }
		                    }
		                ]
		            }
		        ]
		    }
		    `,
				resultPackageRegistrationPages: []resultPackagePage{},
				resultPackageSpec:              "",
				version:                        "",
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "fetchGitRepositoryFromNuget_error_index",

			args: nugetTestArgs{
				inputPackageName:               "nuget-package",
				resultIndex:                    "",
				resultPackageRegistrationIndex: "",
				resultPackageRegistrationPages: []resultPackagePage{},
				resultPackageSpec:              "",
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "fetchGitRepositoryFromNuget_bad_index",

			args: nugetTestArgs{
				inputPackageName:               "nuget-package",
				resultIndex:                    "foo",
				resultPackageRegistrationIndex: "",
				resultPackageRegistrationPages: []resultPackagePage{},
				resultPackageSpec:              "",
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "fetchGitRepositoryFromNuget_error_package_registration_index",

			args: nugetTestArgs{
				inputPackageName: "nuget-package",
				resultIndex: `
		    {
		      "version": "3.0.0",
		      "resources": [
		        {
		          "@id": "https://api.nuget.org/v3-flatcontainer/",
		          "@type": "PackageBaseAddress/3.0.0",
		          "comment": "Base URL of where NuGet packages are stored, in the format ..."
		        },
		        {
		          "@id": "https://api.nuget.org/v3/registration5-gz-semver1/",
		          "@type": "RegistrationsBaseUrl/3.6.0",
		          "comment": "Base URL of Azure storage where NuGet package registration info."
		        }
		        ]
		    }
		    `,
				resultPackageRegistrationIndex: "",
				resultPackageRegistrationPages: []resultPackagePage{},
				resultPackageSpec:              "",
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "fetchGitRepositoryFromNuget_bad_package_index",

			args: nugetTestArgs{
				inputPackageName: "nuget-package",
				resultIndex: `
		    {
		      "version": "3.0.0",
		      "resources": [
		        {
		          "@id": "https://api.nuget.org/v3-flatcontainer/",
		          "@type": "PackageBaseAddress/3.0.0",
		          "comment": "Base URL of where NuGet packages are stored, in the format ..."
		        },
		        {
		          "@id": "https://api.nuget.org/v3/registration5-gz-semver1/",
		          "@type": "RegistrationsBaseUrl/3.6.0",
		          "comment": "Base URL of Azure storage where NuGet package registration info."
		        }
		        ]
		    }
		    `,
				resultPackageRegistrationIndex: "foo",
				resultPackageRegistrationPages: []resultPackagePage{},
				resultPackageSpec:              "",
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "fetchGitRepositoryFromNuget_error_package_registration_page",
			args: nugetTestArgs{
				inputPackageName: "nuget-package",
				resultIndex: `
		    {
		      "version": "3.0.0",
		      "resources": [
		        {
		          "@id": "https://api.nuget.org/v3-flatcontainer/",
		          "@type": "PackageBaseAddress/3.0.0",
		          "comment": "Base URL of where NuGet packages are stored, in the format ..."
		        },
		        {
		          "@id": "https://api.nuget.org/v3/registration5-gz-semver1/",
		          "@type": "RegistrationsBaseUrl/3.6.0",
		          "comment": "Base URL of Azure storage where NuGet package registration info."
		        }
		        ]
		    }
		    `,
				resultPackageRegistrationIndex: `
				{
					"@id": "https://api.nuget.org/v3/registration5-semver1/Foo.NET/index.json",
					"count": 2,
					"items": [
						{
							"@id": "https://api.nuget.org/v3/registration5-semver1/Foo.NET/index.json#page/1",
							"@type": "catalog:CatalogPage",
							"count": 2
						},
						{
							"@id": "https://api.nuget.org/v3/registration5-semver1/Foo.NET/index.json#page/2",
							"@type": "catalog:CatalogPage",
							"count": 2
						}
					]
				}
				`,
				resultPackageRegistrationPages: []resultPackagePage{
					{
						url:      "https://api.nuget.org/v3/registration5-semver1/Foo.NET/index.json#page/1",
						response: "",
					},
					{
						url:      "https://api.nuget.org/v3/registration5-semver1/Foo.NET/index.json#page/2",
						response: "",
					},
				},
				resultPackageSpec: "",
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "fetchGitRepositoryFromNuget_bad_package_registration_page",
			args: nugetTestArgs{
				inputPackageName: "nuget-package",
				resultIndex: `
		    {
		      "version": "3.0.0",
		      "resources": [
		        {
		          "@id": "https://api.nuget.org/v3-flatcontainer/",
		          "@type": "PackageBaseAddress/3.0.0",
		          "comment": "Base URL of where NuGet packages are stored, in the format ..."
		        },
		        {
		          "@id": "https://api.nuget.org/v3/registration5-gz-semver1/",
		          "@type": "RegistrationsBaseUrl/3.6.0",
		          "comment": "Base URL of Azure storage where NuGet package registration info."
		        }
		        ]
		    }
		    `,
				resultPackageRegistrationIndex: `
				{
					"@id": "https://api.nuget.org/v3/registration5-semver1/Foo.NET/index.json",
					"count": 2,
					"items": [
						{
							"@id": "https://api.nuget.org/v3/registration5-semver1/Foo.NET/index.json#page/1",
							"@type": "catalog:CatalogPage",
							"count": 2
						},
						{
							"@id": "https://api.nuget.org/v3/registration5-semver1/Foo.NET/index.json#page/2",
							"@type": "catalog:CatalogPage",
							"count": 2
						}
					]
				}
				`,
				resultPackageRegistrationPages: []resultPackagePage{
					{
						url:      "https://api.nuget.org/v3/registration5-semver1/Foo.NET/index.json#page/1",
						response: "foo",
					},
					{
						url:      "https://api.nuget.org/v3/registration5-semver1/Foo.NET/index.json#page/2",
						response: "foo",
					},
				},
				resultPackageSpec: "",
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "fetchGitRepositoryFromNuget_error_package_spec",
			args: nugetTestArgs{
				inputPackageName: "nuget-package",
				resultIndex: `
		    {
		      "version": "3.0.0",
		      "resources": [
		        {
		          "@id": "https://api.nuget.org/v3-flatcontainer/",
		          "@type": "PackageBaseAddress/3.0.0",
		          "comment": "Base URL of where NuGet packages are stored, in the format ..."
		        },
		        {
		          "@id": "https://api.nuget.org/v3/registration5-gz-semver1/",
		          "@type": "RegistrationsBaseUrl/3.6.0",
		          "comment": "Base URL of Azure storage where NuGet package registration info."
		        }
		        ]
		    }
		    `,
				resultPackageRegistrationIndex: `
		    {
		        "@id": "https://api.nuget.org/v3/registration5-semver1/Foo.NET/index.json",
		        "count": 1,
		        "items": [
		            {
		                "@id": "https://api.nuget.org/v3/registration5-semver1/Foo.NET/index.json#page/1",
		                "@type": "catalog:CatalogPage",
		                "count": 2,
		                "items": [
		                    {
		                        "@id": "https://api.nuget.org/v3/registration5-semver1/Foo.NET/3.5.8.json",
		                        "@type": "Package",
		                        "catalogEntry": {
		                            "@id": "https://api.nuget.org/v3/catalog0/data/2022.12.08.16.43.03/Foo.NET.3.5.8.json",
		                            "@type": "PackageDetails",
		                            "listed": true,
		                            "version": "3.5.8"
		                        }
		                    },
		                    {
		                        "@id": "https://api.nuget.org/v3/registration5-semver1/Foo.NET/4.0.1.json",
		                        "@type": "Package",
		                        "catalogEntry": {
		                            "@id": "https://api.nuget.org/v3/catalog0/data/2022.12.08.16.43.03/Foo.NET.4.0.1.json",
		                            "@type": "PackageDetails",
		                            "listed": true,
		                            "version": "4.0.1"
		                        }
		                    }
		                ]
		            }
		        ]
		    }
		    `,
				resultPackageRegistrationPages: []resultPackagePage{},
				resultPackageSpec:              "",
				version:                        "4.0.1",
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "fetchGitRepositoryFromNuget_bad_package_spec",

			args: nugetTestArgs{
				inputPackageName: "nuget-package",
				resultIndex: `
		    {
		      "version": "3.0.0",
		      "resources": [
		        {
		          "@id": "https://api.nuget.org/v3-flatcontainer/",
		          "@type": "PackageBaseAddress/3.0.0",
		          "comment": "Base URL of where NuGet packages are stored, in the format ..."
		        },
		        {
		          "@id": "https://api.nuget.org/v3/registration5-gz-semver1/",
		          "@type": "RegistrationsBaseUrl/3.6.0",
		          "comment": "Base URL of Azure storage where NuGet package registration info."
		        }
		        ]
		    }
		    `,
				resultPackageRegistrationIndex: `
				{
					"@id": "https://api.nuget.org/v3/registration5-semver1/Foo.NET/index.json",
					"count": 2,
					"items": [
						{
							"@id": "https://api.nuget.org/v3/registration5-semver1/Foo.NET/index.json#page/1",
							"@type": "catalog:CatalogPage",
							"count": 2
						},
						{
							"@id": "https://api.nuget.org/v3/registration5-semver1/Foo.NET/index.json#page/2",
							"@type": "catalog:CatalogPage",
							"count": 2
						}
					]
				}
				`,
				resultPackageRegistrationPages: []resultPackagePage{
					{
						url:      "https://api.nuget.org/v3/registration5-semver1/Foo.NET/index.json#page/1",
						response: "foo",
					},
				},
				resultPackageSpec: "",
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "fetchGitRepositoryFromNuget_error_package_spec",
			args: nugetTestArgs{
				inputPackageName: "nuget-package",
				resultIndex: `
		    {
		      "version": "3.0.0",
		      "resources": [
		        {
		          "@id": "https://api.nuget.org/v3-flatcontainer/",
		          "@type": "PackageBaseAddress/3.0.0",
		          "comment": "Base URL of where NuGet packages are stored, in the format ..."
		        },
		        {
		          "@id": "https://api.nuget.org/v3/registration5-gz-semver1/",
		          "@type": "RegistrationsBaseUrl/3.6.0",
		          "comment": "Base URL of Azure storage where NuGet package registration info."
		        }
		        ]
		    }
		    `,
				resultPackageRegistrationIndex: `
		    {
		        "@id": "https://api.nuget.org/v3/registration5-semver1/Foo.NET/index.json",
		        "count": 1,
		        "items": [
		            {
		                "@id": "https://api.nuget.org/v3/registration5-semver1/Foo.NET/index.json#page/1",
		                "@type": "catalog:CatalogPage",
		                "count": 2,
		                "items": [
		                    {
		                        "@id": "https://api.nuget.org/v3/registration5-semver1/Foo.NET/3.5.8.json",
		                        "@type": "Package",
		                        "catalogEntry": {
		                            "@id": "https://api.nuget.org/v3/catalog0/data/2022.12.08.16.43.03/Foo.NET.3.5.8.json",
		                            "@type": "PackageDetails",
		                            "listed": true,
		                            "version": "3.5.8"
		                        }
		                    },
		                    {
		                        "@id": "https://api.nuget.org/v3/registration5-semver1/Foo.NET/4.0.1.json",
		                        "@type": "Package",
		                        "catalogEntry": {
		                            "@id": "https://api.nuget.org/v3/catalog0/data/2022.12.08.16.43.03/Foo.NET.4.0.1.json",
		                            "@type": "PackageDetails",
		                            "listed": true,
		                            "version": "4.0.1"
		                        }
		                    }
		                ]
		            }
		        ]
		    }
		    `,
				resultPackageRegistrationPages: []resultPackagePage{},
				resultPackageSpec:              "foo",
				version:                        "4.0.1",
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "fetchGitRepositoryFromNuget_missing_repo",

			args: nugetTestArgs{
				inputPackageName: "nuget-package",
				resultIndex: `
		    {
		      "version": "3.0.0",
		      "resources": [
		        {
		          "@id": "https://api.nuget.org/v3-flatcontainer/",
		          "@type": "PackageBaseAddress/3.0.0",
		          "comment": "Base URL of where NuGet packages are stored, in the format ..."
		        },
		        {
		          "@id": "https://api.nuget.org/v3/registration5-gz-semver1/",
		          "@type": "RegistrationsBaseUrl/3.6.0",
		          "comment": "Base URL of Azure storage where NuGet package registration info."
		        }
		        ]
		    }
		    `,
				resultPackageRegistrationIndex: `
				{
					"@id": "https://api.nuget.org/v3/registration5-semver1/Foo.NET/index.json",
					"count": 2,
					"items": [
						{
							"@id": "https://api.nuget.org/v3/registration5-semver1/Foo.NET/index.json#page/1",
							"@type": "catalog:CatalogPage",
							"count": 2
						},
						{
							"@id": "https://api.nuget.org/v3/registration5-semver1/Foo.NET/index.json#page/2",
							"@type": "catalog:CatalogPage",
							"count": 2
						}
					]
				}
				`,
				resultPackageRegistrationPages: []resultPackagePage{
					{
						url:      "https://api.nuget.org/v3/registration5-semver1/Foo.NET/index.json#page/1",
						response: "foo",
					},
				},
				resultPackageSpec: "",
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "fetchGitRepositoryFromNuget_error_no_registration_url",
			args: nugetTestArgs{
				inputPackageName: "nuget-package",
				resultIndex: `
		    {
		      "version": "3.0.0",
		      "resources": [
		        {
		          "@id": "https://api.nuget.org/v3-flatcontainer/",
		          "@type": "PackageBaseAddress/3.0.0",
		          "comment": "Base URL of where NuGet packages are stored, in the format ..."
		        },
		        {
		          "@id": "https://api.nuget.org/v3/registration5-gz-semver1/",
		          "@type": "RegistrationsBaseUrl/3.2.0",
		          "comment": "Base URL of Azure storage where NuGet package registration info."
		        }
		        ]
		    }
		    `,
				resultPackageRegistrationIndex: "",
				resultPackageRegistrationPages: []resultPackagePage{},
				resultPackageSpec:              "",
				version:                        "4.0.1",
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "fetchGitRepositoryFromNuget_error_no_package_base_url",
			args: nugetTestArgs{
				inputPackageName: "nuget-package",
				resultIndex: `
		    {
		      "version": "3.0.0",
		      "resources": [
		        {
		          "@id": "https://api.nuget.org/v3-flatcontainer/",
		          "@type": "PackageBaseAddress/3.1.0",
		          "comment": "Base URL of where NuGet packages are stored, in the format ..."
		        },
		        {
		          "@id": "https://api.nuget.org/v3/registration5-gz-semver1/",
		          "@type": "RegistrationsBaseUrl/3.6.0",
		          "comment": "Base URL of Azure storage where NuGet package registration info."
		        }
		        ]
		    }
		    `,
				resultPackageRegistrationIndex: "",
				resultPackageRegistrationPages: []resultPackagePage{},
				resultPackageSpec:              "",
				version:                        "4.0.1",
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "fetchGitRepositoryFromNuget_error_marhsal_entry",
			args: nugetTestArgs{
				inputPackageName: "nuget-package",
				resultIndex: `
		    {
		      "version": "3.0.0",
		      "resources": [
		        {
		          "@id": "https://api.nuget.org/v3-flatcontainer/",
		          "@type": "PackageBaseAddress/3.0.0",
		          "comment": "Base URL of where NuGet packages are stored, in the format ..."
		        },
		        {
		          "@id": "https://api.nuget.org/v3/registration5-gz-semver1/",
		          "@type": "RegistrationsBaseUrl/3.6.0",
		          "comment": "Base URL of Azure storage where NuGet package registration info."
		        }
		        ]
		    }
		    `,
				resultPackageRegistrationIndex: `
		    {
		        "@id": "https://api.nuget.org/v3/registration5-semver1/Foo.NET/index.json",
		        "count": 1,
		        "items": [
		            {
		                "@id": "https://api.nuget.org/v3/registration5-semver1/Foo.NET/index.json#page/1",
		                "@type": "catalog:CatalogPage",
		                "count": 2,
		                "items": [
		                    {
		                        "@id": "https://api.nuget.org/v3/registration5-semver1/Foo.NET/3.5.8.json",
		                        "@type": "Package",
		                        "catalogEntry": {
		                            "@id": "https://api.nuget.org/v3/catalog0/data/2022.12.08.16.43.03/Foo.NET.3.5.8.json",
		                            "@type": "PackageDetails",
		                            "listed": 123,
								                "version": "3.5.8"
		                        }
		                    },
		                    {
		                        "@id": "https://api.nuget.org/v3/registration5-semver1/Foo.NET/4.0.1.json",
		                        "@type": "Package",
		                        "catalogEntry": {
		                            "@id": "https://api.nuget.org/v3/catalog0/data/2022.12.08.16.43.03/Foo.NET.4.0.1.json",
		                            "@type": "PackageDetails",
		                            "listed": true,
		                            "version": "4.0.1"
		                        }
		                    }
		                ]
		            }
		        ]
		    }
		    `,
				resultPackageRegistrationPages: []resultPackagePage{},
				resultPackageSpec: `
		    <package xmlns="http://schemas.microsoft.com/packaging/2013/05/nuspec.xsd">
		<metadata minClientVersion="2.12">
		      <id>Foo</id>
		      <version>4.0.1</version>
		      <title>Foo.NET</title>
		  </metadata>
		      </package>
		    `,
				version: "4.0.1",
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "fetchGitRepositoryFromNuget_error_package_spec",
			args: nugetTestArgs{
				inputPackageName: "nuget-package",
				resultIndex: `
		    {
		      "version": "3.0.0",
		      "resources": [
		        {
		          "@id": "https://api.nuget.org/v3-flatcontainer/",
		          "@type": "PackageBaseAddress/3.0.0",
		          "comment": "Base URL of where NuGet packages are stored, in the format ..."
		        },
		        {
		          "@id": "https://api.nuget.org/v3/registration5-gz-semver1/",
		          "@type": "RegistrationsBaseUrl/3.6.0",
		          "comment": "Base URL of Azure storage where NuGet package registration info."
		        }
		        ]
		    }
		    `,
				resultPackageRegistrationIndex: `
		    {
		        "@id": "https://api.nuget.org/v3/registration5-semver1/Foo.NET/index.json",
		        "count": 1,
		        "items": [
		            {
		                "@id": "https://api.nuget.org/v3/registration5-semver1/Foo.NET/index.json#page/1",
		                "@type": "catalog:CatalogPage",
		                "count": 2,
		                "items": [
		                    {
		                        "@id": "https://api.nuget.org/v3/registration5-semver1/Foo.NET/3.5.8.json",
		                        "@type": "Package",
		                        "catalogEntry": {
		                            "@id": "https://api.nuget.org/v3/catalog0/data/2022.12.08.16.43.03/Foo.NET.3.5.8.json",
		                            "@type": "PackageDetails",
		                            "listed": true,
		                            "version": "3.5.8"
		                        }
		                    },
		                    {
		                        "@id": "https://api.nuget.org/v3/registration5-semver1/Foo.NET/4.0.1.json",
		                        "@type": "Package",
		                        "catalogEntry": {
		                            "@id": "https://api.nuget.org/v3/catalog0/data/2022.12.08.16.43.03/Foo.NET.4.0.1.json",
		                            "@type": "PackageDetails",
		                            "listed": true,
		                            "version": "4.0.1"
		                        }
		                    }
		                ]
		            }
		        ]
		    }
		    `,
				resultPackageRegistrationPages: []resultPackagePage{},
				resultPackageSpec: `
		    <package xmlns="http://schemas.microsoft.com/packaging/2013/05/nuspec.xsd">
		<metadata minClientVersion="2.12">
		      <id>Foo</id>
		      <version>4.0.1</version>
		      <title>Foo.NET</title>
		  </metadata>
		      </package>
		    `,
				version: "4.0.1",
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
			p := NewMockpackageManagerClient(ctrl)
			p.EXPECT().GetURI(gomock.Any()).
				DoAndReturn(func(url string) (*http.Response, error) {
					return nugetIndexOrPageTestResults(url, &tt)
				}).AnyTimes()
			expectedPackageName := tt.args.expectedPackageName
			if len(strings.TrimSpace(expectedPackageName)) == 0 {
				expectedPackageName = tt.args.inputPackageName
			}

			p.EXPECT().Get(gomock.Any(), expectedPackageName).
				DoAndReturn(func(url, inputPackageName string) (*http.Response, error) {
					return nugetPackageIndexAndSpecResponse(t, url, &tt)
				}).AnyTimes()
			got, err := fetchGitRepositoryFromNuget(tt.args.inputPackageName, p)
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

func nugetIndexOrPageTestResults(url string, test *nugetTest) (*http.Response, error) {
	if url == "https://api.nuget.org/v3/index.json" {
		if test.wantErr && (test.args.resultIndex == "") {
			//nolint
			return nil, errors.New("error")
		}

		return &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(bytes.NewBufferString(test.args.resultIndex)),
		}, nil
	} else {
		urlResponseIndex := slices.IndexFunc(test.args.resultPackageRegistrationPages,
			func(page resultPackagePage) bool { return page.url == url })
		if urlResponseIndex == -1 {
			//nolint
			return nil, errors.New("error")
		}
		page := test.args.resultPackageRegistrationPages[urlResponseIndex]
		if test.wantErr && (page.response == "") {
			//nolint
			return nil, errors.New("error")
		}

		return &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(bytes.NewBufferString(page.response)),
		}, nil
	}
}

func nugetPackageIndexAndSpecResponse(t *testing.T, url string, test *nugetTest) (*http.Response, error) {
	t.Helper()
	if strings.HasSuffix(url, "index.json") {
		if test.wantErr && test.args.resultPackageRegistrationIndex == "" {
			//nolint
			return nil, errors.New("error")
		}
		return &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(bytes.NewBufferString(test.args.resultPackageRegistrationIndex)),
		}, nil
	} else if strings.HasSuffix(url, ".nuspec") {
		if test.wantErr && test.args.resultPackageSpec == "" {
			//nolint
			return nil, errors.New("error")
		}
		if strings.Contains(url, fmt.Sprintf("/%v/", test.args.version)) {
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(bytes.NewBufferString(test.args.resultPackageSpec)),
			}, nil
		}
		t.Errorf("fetchGitRepositoryFromNuget() version = %v, expected version = %v", url, test.args.version)
	}
	//nolint
	return nil, errors.New("error")
}
