import hmac
import hashlib
import json
import urllib.request
import sys

# Configurations
WEBHOOK_URL = "http://localhost:8080/api/wallet/webhook/razorpay"
WEBHOOK_SECRET = "food_court_razorpay_test_secrect"
STUDENT_ID = "STUD123"  # Change this to a student_id that exists in your foodcourt_wallet database
AMOUNT_PAISE = 50000    # Rs. 500.00 (in paise)

payload = {
    "entity": "event",
    "event": "payment.captured",
    "payload": {
        "payment": {
            "entity": {
                "id": "pay_test_payment_123",
                "amount": AMOUNT_PAISE,
                "currency": "INR",
                "status": "captured",
                "method": "upi",
                "email": "student@example.com",
                "contact": "+919876543210",
                "notes": {
                    "studentId": STUDENT_ID
                },
                "created_at": 1785310000
            }
        }
    }
}

body = json.dumps(payload).encode('utf-8')

# Calculate Razorpay signature: HMAC-SHA256 hex digest of raw request body
mac = hmac.new(WEBHOOK_SECRET.encode('utf-8'), body, hashlib.sha256)
signature = mac.hexdigest()

print(f"Sending webhook payload to: {WEBHOOK_URL}")
print(f"Simulating Student ID: {STUDENT_ID}")
print(f"Simulating Amount: Rs. {AMOUNT_PAISE / 100.0}")
print(f"Signature: {signature}")
print("-" * 50)

req = urllib.request.Request(WEBHOOK_URL, data=body, method='POST')
req.add_header('Content-Type', 'application/json')
req.add_header('X-Razorpay-Signature', signature)

try:
    with urllib.request.urlopen(req) as response:
        status_code = response.getcode()
        resp_body = response.read().decode('utf-8')
        print(f"Response Status Code: {status_code}")
        print(f"Response Body: {resp_body}")
except urllib.error.HTTPError as e:
    print(f"HTTP Error: {e.code}")
    print(f"Response Body: {e.read().decode('utf-8')}")
except Exception as e:
    print(f"Error sending request: {e}")

