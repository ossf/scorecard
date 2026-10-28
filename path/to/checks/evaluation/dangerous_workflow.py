# Modified evaluation logic for the Dangerous Workflow check
import re

def evaluate_dangerous_workflow(variable: str) -> bool:
    """
    Evaluates the Dangerous Workflow check for the given variable.

    Args:
        variable (str): The variable to evaluate.

    Returns:
        bool: True if the variable contains an untrusted context pattern, False otherwise.
    """
    # GitHub event context details that may be attacker-controlled.
    # See https://securitylab.github.com/research/github-actions-untrusted-input/
    untrusted_context_pattern = re.compile(
        r".*(issue\.title|issue\.body|pull_request\.title|pull_request\.body|"
        r"comment\.body|review\.body|review_comment\.body|pages.*\.page_name|"
        r"commits.*\.message|head_commit\.message|head_commit\.author\.email|"
        r"head_commit\.author\.name|commits.*\.author\.email|commits.*\.author\.name|"
        r"pull_request\.head\.ref|pull_request\.head\.label|pull_request\.head\.repo\.default_branch|"
        r"issue_comment\.comment\.body|commit_comment\.comment\.body|fork\.forkee\.name).*"
    )

    if "github.head_ref" in variable:
        return True
    return "github.event." in variable and untrusted_context_pattern.search(variable)