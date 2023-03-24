import os
from typing import Any, Dict

import requests


def get_auth_headers(token: str) -> Dict[str, str]:
    return {"Authorization": f"token {token}", "Accept": "application/vnd.github.v3+json"}


def get_url(endpoint: str) -> str:
    return f"https://api.github.com{endpoint}"


def get_code_diff(owner: str, repo_name: str, branch: str) -> Dict[str, Any]:
    url = get_url(f"/repos/{owner}/{repo_name}/compare/main...{branch}")
    response = requests.get(url, headers=get_auth_headers(os.environ["GITHUB_TOKEN"]))
    response.raise_for_status()
    return response.json()["files"]


def request_review(diff: str, api_key: str) -> str:
    headers = {"Authorization": f"Bearer {api_key}"}
    data = {
        "comment": "",
        "model": "text-davinci-002",
        "prompt": f"Review the following code diff:\n\n{diff}",
        "temperature": 0.5,
        "max_tokens": 1024,
        "n": 1,
    }
    response = requests.post("https://api.openai.com/v1/review/requests", headers=headers, json=data)
    response.raise_for_status()
    return response.json()["choices"][0]["text"]


if __name__ == "__main__":
    import argparse

    parser = argparse.ArgumentParser()
    parser.add_argument("-c", "--current-branch", type=str, required=True)
    parser.add_argument("-f", "--diff-format", type=str, default="unified")
    args = parser.parse_args()

    owner, repo_name = os.environ["GITHUB_REPOSITORY"].split("/")
    diff = get_code_diff(owner, repo_name, args.current_branch)
    review = request_review(diff, os.environ["CHATGPT_API_KEY"])
    print(review)
