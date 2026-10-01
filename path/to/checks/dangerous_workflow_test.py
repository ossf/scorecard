# Modified test cases for the `contains_untrusted_context_pattern` function
import unittest
from checks.dangerous_workflow import contains_untrusted_context_pattern

class TestContainsUntrustedContextPattern(unittest.TestCase):
    def test_issue_title(self):
        variable = "github.event.issue.title"
        self.assertTrue(contains_untrusted_context_pattern(variable))

    def test_issue_body(self):
        variable = "github.event.issue.body"
        self.assertTrue(contains_untrusted_context_pattern(variable))

    def test_pull_request_title(self):
        variable = "github.event.pull_request.title"
        self.assertTrue(contains_untrusted_context_pattern(variable))

    def test_pull_request_body(self):
        variable = "github.event.pull_request.body"
        self.assertTrue(contains_untrusted_context_pattern(variable))

    def test_comment_body(self):
        variable = "github.event.issue_comment.comment.body"
        self.assertTrue(contains_untrusted_context_pattern(variable))

    def test_review_body(self):
        variable = "github.event.review.body"
        self.assertTrue(contains_untrusted_context_pattern(variable))

    def test_review_comment_body(self):
        variable = "github.event.review_comment.body"
        self.assertTrue(contains_untrusted_context_pattern(variable))

    def test_pages_page_name(self):
        variable = "github.event.pages.page_name"
        self.assertTrue(contains_untrusted_context_pattern(variable))

    def test_commits_message(self):
        variable = "github.event.commits.message"
        self.assertTrue(contains_untrusted_context_pattern(variable))

    def test_head_commit_message(self):
        variable = "github.event.head_commit.message"
        self.assertTrue(contains_untrusted_context_pattern(variable))

    def test_head_commit_author_email(self):
        variable = "github.event.head_commit.author.email"
        self.assertTrue(contains_untrusted_context_pattern(variable))

    def test_head_commit_author_name(self):
        variable = "github.event.head_commit.author.name"
        self.assertTrue(contains_untrusted_context_pattern(variable))

    def test_commits_author_email(self):
        variable = "github.event.commits.author.email"
        self.assertTrue(contains_untrusted_context_pattern(variable))

    def test_commits_author_name(self):
        variable = "github.event.commits.author.name"
        self.assertTrue(contains_untrusted_context_pattern(variable))

    def test_pull_request_head_ref(self):
        variable = "github.event.pull_request.head.ref"
        self.assertTrue(contains_untrusted_context_pattern(variable))

    def test_pull_request_head_label(self):
        variable = "github.event.pull_request.head.label"
        self.assertTrue(contains_untrusted_context_pattern(variable))

    def test_pull_request_head_repo_default_branch(self):
        variable = "github.event.pull_request.head.repo.default_branch"
        self.assertTrue(contains_untrusted_context_pattern(variable))

    def test_fork_forkee_name(self):
        variable = "github.event.fork.forkee.name"
        self.assertTrue(contains_untrusted_context_pattern(variable))

    def test_invalid_variable(self):
        variable = "github.event.invalid.variable"
        self.assertFalse(contains_untrusted_context_pattern(variable))

if __name__ == "__main__":
    unittest.main()