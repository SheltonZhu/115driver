import requests


def get_code_diff(branch):
    """Gets the code difference between main branch and current branch"""
    url = f"https://api.github.com/repos/{GITHUB_REPOSITORY}/compare/main...{branch}"
    headers = {
        "Authorization": f"token {GITHUB_TOKEN}",
        "Accept": "application/vnd.github.v3+json",
    }
    response = requests.get(url, headers=headers)
    if response.status_code != 200:
        raise Exception(f"Failed to get diff: {response.content}")
    return response.json()["files"]


def request_review(diff):
    """Requests a review from ChatGPT"""
    url = "https://api.openai.com/v1/review/requests"
    headers = {"Authorization": f"Bearer {CHATGPT_API_KEY}"}
    data = {
        "comment": "",
        "model": "text-davinci-002",
        "prompt": f"Review the following code diff:\n\n{diff}",
        "temperature": 0.5,
        "max_tokens": 1024,
        "n": 1,
    }
    response = requests.post(url, headers=headers, json=data)
    if response.status_code != 200:
        raise Exception(f"Failed to request review: {response.content}")
    return response.json()["choices"][0]["text"]


if __name__ == "__main__":
    import argparse

    parser = argparse.ArgumentParser()
    parser.add_argument("-c", "--current-branch", type=str, required=True)
    parser.add_argument("-f", "--diff-format", type=str, default="unified")
    args = parser.parse_args()

    diff = get_code_diff(args.current_branch)
    review = request_review(diff)
    print(review)
