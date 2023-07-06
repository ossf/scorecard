# OpenSSF Scorecard Contributor ladder

This document outlines the various contributor roles in the Scorecard project, along with their respective prerequisites and responsibilities.
It also defines the process by which users can request to change roles.

- [Roles](#roles)
  - [Community participants](#community-participants)
  - [Community members](#community-members)
  - [Triagers](#triagers)
  - [Maintainers](#maintainers)
- [Inactive members](#inactive-members)

## Roles

### Community participants

Community participants engage with Scorecard,
contributing their time and energy in discussions or just generally helping out.

#### Pre-requisites

- Must follow the [OpenSSF Code of Conduct]
- Must follow the [Contribution Guide]

#### Responsibilities

- Keep it up!

### Community Members

Community Members are active contributors in the community.
They can have issues and PRs assigned to them, participate through GitHub teams,
and pre-submit tests are automatically run for their PRs.
Members are expected to remain active contributors to the community.

**Defined by:** Member of the OpenSSF GitHub organization

#### Pre-requisites

- Enabled two-factor authentication on their GitHub account
- Have made multiple contributions to the project or community.
  Contributions may include, but are not limited to:
  - Authoring or reviewing PRs on GitHub. At least one PR must be **merged**.
  - Filing or commenting on issues on GitHub
  - Contributing to a project, or community discussions (e.g. meetings, Slack,
    email discussion forums, Stack Overflow)
- Active contributor to Scorecard or a relevant OpenSSF SIG

#### Responsibilities

- Can be assigned issues and PRs
- Responsive to issues and PRs assigned to them
- Others can ask for reviews with a `/cc @username`.
- Responsive to mentions of teams they are members of
- Active owner of code they have contributed (unless ownership is explicitly transferred)
  - Ensures code is well tested and that tests consistently pass
  - Addresses bugs or issues discovered after code is accepted

#### Privileges

- Tests run against their PRs automatically

#### Promotion process

- Sponsored by 2 maintainers or triagers. **Note the following requirements for sponsors**:
  - Sponsors must have close interactions with the prospective Member – e.g. 
    code/design/proposal review, coordinating on issues, etc.
  - Sponsors must be reviewers or approvers in at least one CODEOWNERS file.
  - Sponsors should preferably be from multiple OpenSSF member companies to incentivize community integration.
- Open an issue in the project's repository
  - Ensure your sponsors are `@mentioned`
  - Describe and/or link to all your relevant contributions to the project
    (or other OpenSSF projects)
  - Sponsoring reviewers must comment on the issue/PR confirming their sponsorship

### Triagers

Triagers help a project by reviewing issues and code for quality and correctness.
They are knowledgeable about the project's codebase (in its entirety or a specific section)
and software engineering principles.

**Defined by:** "triage" permission in the project

#### Pre-requisites

- Community Member for at least 3 months
- Helped to triage issues and pull requests
- Knowledgeable about the codebase

#### Responsibilities

- Read through issues and PRs
  - Answer questions when possible
  - Add relevant labels
  - Draw maintainers' attention (via `@mention`) if relevant
  - Close issue (as "completed" or "not planned") if necessary
- Help maintain project quality control via [code reviews] on PRs
  - Focus on code quality and correctness, including testing and factoring
  - May also review for more holistic issues, but not a requirement
- Be responsive to review requests
- May be assigned PRs to review if in area of expertise
- Assigned test bugs related to the project of expertise

#### Privileges

- Same as for Community Members
- Triager status may be a precondition to accepting large code contributions

#### Promotion process

- Sponsored by a maintainer
  - With no objections from other maintainers
  - Done through issue or PR to update the CODEOWNERS file
- May self-nominate or be nominated by a maintainer
  - In case of self-nomination, sponsor must comment approval on the issue/PR

### Maintainers

Maintainers are responsible for the project's overall health.
They are the only ones who can approve and merge code contributions.
While triage and code review is focused on code quality and correctness,
approval is focused on holistic acceptance of a contribution including:

- backwards/forwards compatibility
- adherence to API and flag conventions
- subtle performance and correctness issues
- interactions with other parts of the system
- consistency between code and documentation

**Defined by:** "Maintain" permissions in the project and an entry in its CODEOWNERS file

#### Pre-requisites

- Triager for at least 3 months
- Reviewed at least 10 substantial PRs to the codebase
- Reviewed or got at least 30 PRs merged to the codebase

#### Responsibilities

- Demonstrate sound technical judgment
- Maintain project quality control via code reviews
  - Focus on holistic acceptance of contribution
- Be responsive to review requests
- Mentor contributors and triagers
- Approve and merge code contributions as appropriate
- Participate in OpenSSF or Scorecard-specific community meetings, if possible
- Hosting Scorecard-specific community meetings, if possible and comfortable

#### Privileges

- Same as for Triager
- Maintainer status may be a precondition to accepting especially large code contributions

#### Promotion process
- Sponsored by a maintainer
  - With no objections from other maintainers
  - Done through PR to update the CODEOWNERS file
- May self-nominate or be nominated by a maintainer
  - In case of self-nomination, sponsor must comment approval on the PR

## Inactive members
A core principle in maintaining a healthy community is encouraging active participation.
It is inevitable that a contributor's focus will change over time
and there is no expectation they'll actively contribute forever.

Any contributor at any level described above may write an issue (or PR, if CODEOWNER changes are necessary)
asking to step down to a lighter-weight tier or to depart the project entirely.
Such requests will hopefully come after thoughtful conversations with the rest of the team
and with sufficient forewarning for the others to prepare. However, sometimes "life happens".
Therefore, the change in responsibilities will be understood to take immediate effect,
regardless of whether the issue/PR has been acknowledged or merged.

However, should a Triager or above be deemed inactive for a significant period, any
Community Member or above may write an issue/PR requesting their removal from the ranks
(and `@mentioning` the inactive contributor in the hopes of drawing their attention).
The request must receive support (in comments) from a majority of Maintainers to proceed.


[OpenSSF Code of Conduct]: https://openssf.org/community/code-of-conduct/
[Contribution Guide]: ./CONTRIBUTING.md
