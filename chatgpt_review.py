import os
import requests
import json
import sys

def get_chatgpt_review(api_token, pr_number):
    headers = {
      'Authorization': 'Bearer {0}'.format(api_token),
      'Content-Type': 'application/json'
    }
    
    data = {
        'type': 'pull_request',
        'number': pr_number,
        'repository': {
            'full_name': os.environ.get('GITHUB_REPOSITORY')
        }
    }

    response = requests.post('https://api.chatgpt.com/review', headers=headers, json=data)

    return response.json()

def post_to_comment(comment_text, pr_number, github_token):
    url = f'https://api.github.com/repos/{os.environ.get("GITHUB_REPOSITORY")}/issues/{pr_number}/comments'
    payload = {"body": comment_text}
    headers = {"Authorization": f"Bearer {github_token}"}

    requests.post(url, headers=headers, json=payload)

pr_number = int(sys.argv[1])
github_token = sys.argv[2]
chatgpt_api_token = sys.argv[3]

review_result = get_chatgpt_review(chatgpt_api_token, pr_number)
post_to_comment(review_result['data']['comment'], pr_number, github_token)