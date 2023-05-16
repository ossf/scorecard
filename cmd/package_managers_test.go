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
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
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
			p := NewMockpackageManagerClient(ctrl)
			p.EXPECT().Get(gomock.Any(), tt.args.packageName).
				DoAndReturn(func(url, packageName string) (*http.Response, error) {
					if tt.wantErr && tt.args.result == "" {
						//nolint
						return nil, errors.New("error")
					}

					return &http.Response{
						StatusCode: 200,
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
			//nolint
			args: args{
				packageName: "npm-package",
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
			//nolint
			args: args{
				packageName: "npm-package",
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
			p.EXPECT().Get(gomock.Any(), tt.args.packageName).
				DoAndReturn(func(url, packageName string) (*http.Response, error) {
					if tt.wantErr && tt.args.result == "" {
						//nolint
						return nil, errors.New("error")
					}

					return &http.Response{
						StatusCode: 200,
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
			//nolint
			args: args{
				packageName: "npm-package",
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
			//nolint
			args: args{
				packageName: "npm-package",
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
			p.EXPECT().Get(gomock.Any(), tt.args.packageName).
				DoAndReturn(func(url, packageName string) (*http.Response, error) {
					if tt.wantErr && tt.args.result == "" {
						//nolint
						return nil, errors.New("error")
					}

					return &http.Response{
						StatusCode: 200,
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
		packageName        string
		resultIndex        string
		resultPackageIndex string
		resultPackageSpec  string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "fetchGitRepositoryFromNuget",

			args: args{
				packageName: "nuget-package",
				resultIndex: `
        {
          "version": "3.0.0",
          "resources": [
            {
              "@id": "https://azuresearch-usnc.nuget.org/query",
              "@type": "SearchQueryService",
              "comment": "Query endpoint of NuGet Search service (primary)"
            },
            {
              "@id": "https://azuresearch-ussc.nuget.org/query",
              "@type": "SearchQueryService",
              "comment": "Query endpoint of NuGet Search service (secondary)"
            },
            {
              "@id": "https://azuresearch-usnc.nuget.org/autocomplete",
              "@type": "SearchAutocompleteService",
              "comment": "Autocomplete endpoint of NuGet Search service (primary)"
            },
            {
              "@id": "https://api.nuget.org/v3/registration5-semver1/",
              "@type": "RegistrationsBaseUrl",
              "comment": "Base URL of Azure storage where NuGet package registration info is stored"
            },
            {
              "@id": "https://api.nuget.org/v3-flatcontainer/",
              "@type": "PackageBaseAddress/3.0.0",
              "comment": "Base URL of where NuGet packages are stored, in the format ..."
            },
            {
              "@id": "https://www.nuget.org/api/v2",
              "@type": "LegacyGallery"
            },
            {
              "@id": "https://www.nuget.org/api/v2",
              "@type": "LegacyGallery/2.0.0"
            },
            {
              "@id": "https://www.nuget.org/api/v2/package",
              "@type": "PackagePublish/2.0.0"
            }
            ]
        }
        `,
				resultPackageIndex: `
        {
          "versions": [
            "13.0.2",
            "13.0.3-beta1",
            "13.0.3"
          ]
        }
        `,
				resultPackageSpec: `
        <package xmlns="http://schemas.microsoft.com/packaging/2013/05/nuspec.xsd">
          <metadata minClientVersion="2.12">
          <id>Foo</id>
          <version>13.0.3</version>
          <title>Foo.NET</title>
          <authors>Foo Foo</authors>
          <owners>Foo Foo</owners>
          <requireLicenseAcceptance>false</requireLicenseAcceptance>
          <license type="expression">MIT</license>
          <licenseUrl>https://licenses.nuget.org/MIT</licenseUrl>
          <projectUrl>https://www.newtonsoft.com/json</projectUrl>
          <iconUrl>https://www.foo.com/content/images/nugeticon.png</iconUrl>
          <description>Foo.NET is a popular foo framework for .NET</description>
          <copyright>Copyright ©Foo Foo 2008</copyright>
          <tags>foo</tags>
          <repository type="git" url="foo" commit="123"/>
          <dependencies>
          <group targetFramework=".NETFramework2.0"/>
          <group targetFramework=".NETFramework3.5"/>
          <group targetFramework=".NETFramework4.0"/>
          <group targetFramework=".NETFramework4.5"/>
          <group targetFramework=".NETPortable0.0-Profile259"/>
          <group targetFramework=".NETPortable0.0-Profile328"/>
          <group targetFramework=".NETStandard1.0">
          <dependency id="Microsoft.CSharp" version="4.3.0" exclude="Build,Analyzers"/>
          <dependency id="NETStandard.Library" version="1.6.1" exclude="Build,Analyzers"/>
          <dependency id="System.ComponentModel.TypeConverter" version="4.3.0" exclude="Build,Analyzers"/>
          <dependency id="System.Runtime.Serialization.Primitives" version="4.3.0" exclude="Build,Analyzers"/>
          </group>
          <group targetFramework=".NETStandard1.3">
          <dependency id="Microsoft.CSharp" version="4.3.0" exclude="Build,Analyzers"/>
          <dependency id="NETStandard.Library" version="1.6.1" exclude="Build,Analyzers"/>
          <dependency id="System.ComponentModel.TypeConverter" version="4.3.0" exclude="Build,Analyzers"/>
          <dependency id="System.Runtime.Serialization.Formatters" version="4.3.0" exclude="Build,Analyzers"/>
          <dependency id="System.Runtime.Serialization.Primitives" version="4.3.0" exclude="Build,Analyzers"/>
          <dependency id="System.Xml.XmlDocument" version="4.3.0" exclude="Build,Analyzers"/>
          </group>
          <group targetFramework=".NETStandard2.0"/>
          </dependencies>
          </metadata>
          </package>
        `,
			},
			want:    "foo",
			wantErr: false,
		},
		{
			name: "fetchGitRepositoryFromNuget_error_index",

			args: args{
				packageName:        "nuget-package",
				resultIndex:        "",
				resultPackageIndex: "",
				resultPackageSpec:  "",
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "fetchGitRepositoryFromNuget_bad_index",

			args: args{
				packageName:        "nuget-package",
				resultIndex:        "foo",
				resultPackageIndex: "",
				resultPackageSpec:  "",
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "fetchGitRepositoryFromNuget_error_package_index",

			args: args{
				packageName: "nuget-package",
				resultIndex: `
        {
          "version": "3.0.0",
          "resources": [
            {
              "@id": "https://azuresearch-usnc.nuget.org/query",
              "@type": "SearchQueryService",
              "comment": "Query endpoint of NuGet Search service (primary)"
            },
            {
              "@id": "https://azuresearch-ussc.nuget.org/query",
              "@type": "SearchQueryService",
              "comment": "Query endpoint of NuGet Search service (secondary)"
            },
            {
              "@id": "https://azuresearch-usnc.nuget.org/autocomplete",
              "@type": "SearchAutocompleteService",
              "comment": "Autocomplete endpoint of NuGet Search service (primary)"
            },
            {
              "@id": "https://api.nuget.org/v3/registration5-semver1/",
              "@type": "RegistrationsBaseUrl",
              "comment": "Base URL of Azure storage where NuGet package registration info is stored"
            },
            {
              "@id": "https://api.nuget.org/v3-flatcontainer/",
              "@type": "PackageBaseAddress/3.0.0",
              "comment": "Base URL of where NuGet packages are stored, in the format ..."
            },
            {
              "@id": "https://www.nuget.org/api/v2",
              "@type": "LegacyGallery"
            },
            {
              "@id": "https://www.nuget.org/api/v2",
              "@type": "LegacyGallery/2.0.0"
            },
            {
              "@id": "https://www.nuget.org/api/v2/package",
              "@type": "PackagePublish/2.0.0"
            }
            ]
        }
        `,
				resultPackageIndex: "",
				resultPackageSpec:  "",
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "fetchGitRepositoryFromNuget_bad_package_index",

			args: args{
				packageName: "nuget-package",
				resultIndex: `
        {
          "version": "3.0.0",
          "resources": [
            {
              "@id": "https://azuresearch-usnc.nuget.org/query",
              "@type": "SearchQueryService",
              "comment": "Query endpoint of NuGet Search service (primary)"
            },
            {
              "@id": "https://azuresearch-ussc.nuget.org/query",
              "@type": "SearchQueryService",
              "comment": "Query endpoint of NuGet Search service (secondary)"
            },
            {
              "@id": "https://azuresearch-usnc.nuget.org/autocomplete",
              "@type": "SearchAutocompleteService",
              "comment": "Autocomplete endpoint of NuGet Search service (primary)"
            },
            {
              "@id": "https://api.nuget.org/v3/registration5-semver1/",
              "@type": "RegistrationsBaseUrl",
              "comment": "Base URL of Azure storage where NuGet package registration info is stored"
            },
            {
              "@id": "https://api.nuget.org/v3-flatcontainer/",
              "@type": "PackageBaseAddress/3.0.0",
              "comment": "Base URL of where NuGet packages are stored, in the format ..."
            },
            {
              "@id": "https://www.nuget.org/api/v2",
              "@type": "LegacyGallery"
            },
            {
              "@id": "https://www.nuget.org/api/v2",
              "@type": "LegacyGallery/2.0.0"
            },
            {
              "@id": "https://www.nuget.org/api/v2/package",
              "@type": "PackagePublish/2.0.0"
            }
            ]
        }
        `,
				resultPackageIndex: "foo",
				resultPackageSpec:  "",
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "fetchGitRepositoryFromNuget_error_package_spec",

			args: args{
				packageName: "nuget-package",
				resultIndex: `
        {
          "version": "3.0.0",
          "resources": [
            {
              "@id": "https://azuresearch-usnc.nuget.org/query",
              "@type": "SearchQueryService",
              "comment": "Query endpoint of NuGet Search service (primary)"
            },
            {
              "@id": "https://azuresearch-ussc.nuget.org/query",
              "@type": "SearchQueryService",
              "comment": "Query endpoint of NuGet Search service (secondary)"
            },
            {
              "@id": "https://azuresearch-usnc.nuget.org/autocomplete",
              "@type": "SearchAutocompleteService",
              "comment": "Autocomplete endpoint of NuGet Search service (primary)"
            },
            {
              "@id": "https://api.nuget.org/v3/registration5-semver1/",
              "@type": "RegistrationsBaseUrl",
              "comment": "Base URL of Azure storage where NuGet package registration info is stored"
            },
            {
              "@id": "https://api.nuget.org/v3-flatcontainer/",
              "@type": "PackageBaseAddress/3.0.0",
              "comment": "Base URL of where NuGet packages are stored, in the format ..."
            },
            {
              "@id": "https://www.nuget.org/api/v2",
              "@type": "LegacyGallery"
            },
            {
              "@id": "https://www.nuget.org/api/v2",
              "@type": "LegacyGallery/2.0.0"
            },
            {
              "@id": "https://www.nuget.org/api/v2/package",
              "@type": "PackagePublish/2.0.0"
            }
            ]
        }
        `,
				resultPackageIndex: `
        {
          "versions": [
            "13.0.2",
            "13.0.3-beta1",
            "13.0.3"
          ]
        }
        `,
				resultPackageSpec: "",
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "fetchGitRepositoryFromNuget_bad_package_spec",

			args: args{
				packageName: "nuget-package",
				resultIndex: `
        {
          "version": "3.0.0",
          "resources": [
            {
              "@id": "https://azuresearch-usnc.nuget.org/query",
              "@type": "SearchQueryService",
              "comment": "Query endpoint of NuGet Search service (primary)"
            },
            {
              "@id": "https://azuresearch-ussc.nuget.org/query",
              "@type": "SearchQueryService",
              "comment": "Query endpoint of NuGet Search service (secondary)"
            },
            {
              "@id": "https://azuresearch-usnc.nuget.org/autocomplete",
              "@type": "SearchAutocompleteService",
              "comment": "Autocomplete endpoint of NuGet Search service (primary)"
            },
            {
              "@id": "https://api.nuget.org/v3/registration5-semver1/",
              "@type": "RegistrationsBaseUrl",
              "comment": "Base URL of Azure storage where NuGet package registration info is stored"
            },
            {
              "@id": "https://api.nuget.org/v3-flatcontainer/",
              "@type": "PackageBaseAddress/3.0.0",
              "comment": "Base URL of where NuGet packages are stored, in the format ..."
            },
            {
              "@id": "https://www.nuget.org/api/v2",
              "@type": "LegacyGallery"
            },
            {
              "@id": "https://www.nuget.org/api/v2",
              "@type": "LegacyGallery/2.0.0"
            },
            {
              "@id": "https://www.nuget.org/api/v2/package",
              "@type": "PackagePublish/2.0.0"
            }
            ]
        }
        `,
				resultPackageIndex: `
        {
          "versions": [
            "13.0.2",
            "13.0.3-beta1",
            "13.0.3"
          ]
        }
        `,
				resultPackageSpec: "foo",
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "fetchGitRepositoryFromNuget_missing_repo",

			args: args{
				packageName: "nuget-package",
				resultIndex: `
        {
          "version": "3.0.0",
          "resources": [
            {
              "@id": "https://azuresearch-usnc.nuget.org/query",
              "@type": "SearchQueryService",
              "comment": "Query endpoint of NuGet Search service (primary)"
            },
            {
              "@id": "https://azuresearch-ussc.nuget.org/query",
              "@type": "SearchQueryService",
              "comment": "Query endpoint of NuGet Search service (secondary)"
            },
            {
              "@id": "https://azuresearch-usnc.nuget.org/autocomplete",
              "@type": "SearchAutocompleteService",
              "comment": "Autocomplete endpoint of NuGet Search service (primary)"
            },
            {
              "@id": "https://api.nuget.org/v3/registration5-semver1/",
              "@type": "RegistrationsBaseUrl",
              "comment": "Base URL of Azure storage where NuGet package registration info is stored"
            },
            {
              "@id": "https://api.nuget.org/v3-flatcontainer/",
              "@type": "PackageBaseAddress/3.0.0",
              "comment": "Base URL of where NuGet packages are stored, in the format ..."
            },
            {
              "@id": "https://www.nuget.org/api/v2",
              "@type": "LegacyGallery"
            },
            {
              "@id": "https://www.nuget.org/api/v2",
              "@type": "LegacyGallery/2.0.0"
            },
            {
              "@id": "https://www.nuget.org/api/v2/package",
              "@type": "PackagePublish/2.0.0"
            }
            ]
        }
        `,
				resultPackageIndex: `
        {
          "versions": [
            "13.0.2",
            "13.0.3-beta1",
            "13.0.3"
          ]
        }
        `,
				resultPackageSpec: `
        <package xmlns="http://schemas.microsoft.com/packaging/2013/05/nuspec.xsd">
          <metadata minClientVersion="2.12">
          <id>Foo</id>
          <version>13.0.3</version>
          <title>Foo.NET</title>
          <authors>Foo Foo</authors>
          <owners>Foo Foo</owners>
          <requireLicenseAcceptance>false</requireLicenseAcceptance>
          <license type="expression">MIT</license>
          <licenseUrl>https://licenses.nuget.org/MIT</licenseUrl>
          <projectUrl>https://www.newtonsoft.com/json</projectUrl>
          <iconUrl>https://www.foo.com/content/images/nugeticon.png</iconUrl>
          <description>Foo.NET is a popular foo framework for .NET</description>
          <copyright>Copyright ©Foo Foo 2008</copyright>
          <tags>foo</tags>
          <dependencies>
          <group targetFramework=".NETFramework2.0"/>
          <group targetFramework=".NETFramework3.5"/>
          <group targetFramework=".NETFramework4.0"/>
          <group targetFramework=".NETFramework4.5"/>
          <group targetFramework=".NETPortable0.0-Profile259"/>
          <group targetFramework=".NETPortable0.0-Profile328"/>
          <group targetFramework=".NETStandard1.0">
          <dependency id="Microsoft.CSharp" version="4.3.0" exclude="Build,Analyzers"/>
          <dependency id="NETStandard.Library" version="1.6.1" exclude="Build,Analyzers"/>
          <dependency id="System.ComponentModel.TypeConverter" version="4.3.0" exclude="Build,Analyzers"/>
          <dependency id="System.Runtime.Serialization.Primitives" version="4.3.0" exclude="Build,Analyzers"/>
          </group>
          <group targetFramework=".NETStandard1.3">
          <dependency id="Microsoft.CSharp" version="4.3.0" exclude="Build,Analyzers"/>
          <dependency id="NETStandard.Library" version="1.6.1" exclude="Build,Analyzers"/>
          <dependency id="System.ComponentModel.TypeConverter" version="4.3.0" exclude="Build,Analyzers"/>
          <dependency id="System.Runtime.Serialization.Formatters" version="4.3.0" exclude="Build,Analyzers"/>
          <dependency id="System.Runtime.Serialization.Primitives" version="4.3.0" exclude="Build,Analyzers"/>
          <dependency id="System.Xml.XmlDocument" version="4.3.0" exclude="Build,Analyzers"/>
          </group>
          <group targetFramework=".NETStandard2.0"/>
          </dependencies>
          </metadata>
          </package>
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
			p.EXPECT().GetURI(gomock.Any()).
				DoAndReturn(func(url string) (*http.Response, error) {
					if tt.wantErr && tt.args.resultIndex == "" {
						//nolint
						return nil, errors.New("error")
					}

					return &http.Response{
						StatusCode: 200,
						Body:       io.NopCloser(bytes.NewBufferString(tt.args.resultIndex)),
					}, nil
				}).AnyTimes()
			p.EXPECT().Get(gomock.Any(), tt.args.packageName).
				DoAndReturn(func(url, packageName string) (*http.Response, error) {
					if tt.wantErr && tt.args.resultPackageIndex == "" {
						//nolint
						return nil, errors.New("error")
					}

					if tt.wantErr && tt.args.resultPackageSpec == "" {
						//nolint
						return nil, errors.New("error")
					}
					if strings.HasSuffix(url, "index.json") {
						return &http.Response{
							StatusCode: 200,
							Body:       io.NopCloser(bytes.NewBufferString(tt.args.resultPackageIndex)),
						}, nil
					} else if strings.HasSuffix(url, ".nuspec") {
						return &http.Response{
							StatusCode: 200,
							Body:       io.NopCloser(bytes.NewBufferString(tt.args.resultPackageSpec)),
						}, nil
					}
					return nil, nil
				}).AnyTimes()
			got, err := fetchGitRepositoryFromNuget(tt.args.packageName, p)
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
