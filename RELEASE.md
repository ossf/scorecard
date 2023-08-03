# Releasing Scorecard

This is a draft document to describe the release process for Scorecard

(If there are improvements you'd like to see, please comment on the
[tracking issue](https://github.com/ossf/scorecard/issues/1676) or issue a
pull request to discuss.)

- [Tracking](#tracking)
- [Preparing the release](#preparing-the-release)
  - [Validate tests](#validate-tests)
- [Drafting release notes](#drafting-release-notes)
- [Release](#release)
  - [Create a tag](#create-a-tag)
  - [Create a GitHub release](#create-a-github-release)
- [Validate Release](#validate-release)

## Tracking

As the first task, a Release Manager should open a tracking issue for the
release.

We don't currently have a template for releasing, but the following
[issue](https://github.com/ossf/scorecard-action/issues/97) is a good example
to draw inspiration from.

We're not striving for perfection with the template, but the tracking issue
will serve as a reference point to aggregate feedback, so try your best to be
as descriptive as possible.

## Preparing the release

This section covers changes that need to be issued as a pull request and should
be merged before releasing the scorecard GitHub Action.

### Validate tests

Check the unit tests and integration tests are passing for the planned release commit, either locally or for the GitHub workflows.

## Drafting release notes

Release notes are a semi-automated process. We often start by opening [drafting a new release on GitHub](https://github.com/ossf/scorecard/releases/new).
You can select to create a new tag on publish, and auto-generate some notes by clicking `Generate release notes`.
This provides a good start, but no one wants to see a wave of dependabot commits, so filter them out.
Try to focus on the PRs that affect users or behavior, not dependency updates or CI changes.

Using the Kubernetes `release-notes` tool can also be helpful if PR authors filled out the user-facing change section.
```console
release-notes --org ossf --repo scorecard --branch main \
  --dependencies=false \
  --required-author "" \
  --start-rev <previous release tag> \
  --end-rev <commit to be tagged>
```

Note: This doesn't always grab the right value when PR bodies have multiple code blocks in them.

Save your draft when satisfied and share it with other maintainers for feedback, if possible.

## Release

### Create a tag

The GitHub release process supports creating a tag on publish, but prefer signing the tag when possible.
In this example, we're releasing a hypothetical `v100.0.0` at the desired commit SHA `SHA`:

```console
git remote update
git checkout `SHA`
git tag -s -m "v100.0.0" v100.0.0
git push <upstream> v100.0.0
```

### Create a GitHub release

Revisit the draft release you created earlier, and ensure it's using the correct tag.

Release title: `<tag>`

The release notes will be the notes you drafted in the previous step.

Ensure the release is marked as the latest release, if appropriate.

Click `Publish release`.

## Validate Release

When a new tag is pushed, our GitHub Actions will create a release using `goreleaser`.
Confirm the workflow ran without issues. Check the release again to verify the artifacts and provenance have been added.

If any issues were encountered, fixes must be issued under a new release/tag as Go releases are immutable.
