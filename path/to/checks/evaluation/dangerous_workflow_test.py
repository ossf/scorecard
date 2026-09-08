# Modified test cases for the evaluation logic
import unittest
from checks.evaluation.dangerous_workflow import evaluate_dangerous_workflow

class TestEvaluateDangerousWorkflow(unittest.TestCase):
    def test_issue_title(self):
        variable = "github.event.issue.title"
        self.assertTrue(evaluate_dangerous_workflow(variable))

    def test_issue_body(self):
        variable = "github.event.issue.body"
        self.assertTrue(evaluate_dangerous_workflow(variable))

    def test_pull_request_title(self):
        variable = "github.event.pull_request.title"
        self.assertTrue(evaluate_dangerous_workflow(variable))

    def test_pull_request_body(self):
        variable = "github.event.pull_request.body"
        self.assertTrue(evaluate_dangerous_workflow(variable))

    def test_comment_body(self):
        variable = "github.event.issue_comment.comment.body"
        self.assertTrue(evaluate_dangerous_workflow(variable))

    def test_review_body(self):
        variable = "github.event.review.body"
        self.assertTrue(evaluate_dangerous_workflow(variable))

    def test_review_comment_body(self):
        variable = "github.event.review_comment.body"
        self.assertTrue(evaluate_dangerous_workflow(variable))

    def test_pages_page_name(self):
        variable = "github.event.pages.page_name"
        self.assertTrue(evaluate_dangerous_workflow(variable))

    def test_commits_message(self):
        variable = "github.event.commits.message"
        self.assertTrue(evaluate_dangerous_workflow(variable))

    def test_head_commit_message(self):
        variable = "github.event.head_commit.message"
        self.assertTrue(evaluate_dangerous_workflow(variable))

    def test_head_commit_author_email(self):
        variable = "github.event.head_commit.author.email"
        self.assertTrue(evaluate_dangerous_workflow(variable))

    def test_head_commit_author_name(self):
        variable = "github.event.head_commit.author.name"
        self.assertTrue(evaluate_dangerous_workflow(variable))

    def test_commits_author_email(self):
        variable = "github.event.commits.author.email"
        self.assertTrue(evaluate_dangerous_workflow(variable))

    def test_commits_author_name(self):
        variable = "github.event.commits.author.name"
        self.assertTrue(evaluate_dangerous_workflow(variable))

    def test_pull_request_head_ref(self):
        variable = "github.event.pull_request.head.ref"
        self.assertTrue(evaluate_dangerous_workflow(variable))

    def test_pull_request_head_label(self):
        variable = "github.event.pull_request.head.label"
        self.assertTrue(evaluate_dangerous_workflow(variable))

    def test_pull_request_head_repo_default_branch(self):
        variable = "github.event.pull_request.head.repo.default_branch"
        self.assertTrue(evaluate_dangerous_workflow(variable))

    def test_fork_forkee_name(self):
        variable = "github.event.fork.forkee.name"
        self.assertTrue(evaluate_dangerous_workflow(variable))

    def test_invalid_variable(self):
        variable = "github.event.invalid.variable"
        self.assertFalse(evaluate_dangerous_workflow(variable))

if __name__ == "__main__":
    unittest.main()