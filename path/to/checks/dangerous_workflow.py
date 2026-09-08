# Modified version of the original function to include additional GitHub event context details
import re

def contains_untrusted_context_pattern(variable: str) -> bool:
    """
    Checks if the given variable contains any untrusted context patterns.

    Args:
        variable (str): The variable to check.

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