```
docker buildx build --platform linux/amd64 -t evylang/learn:v0.0.5 .
docker push evylang/learn:v0.0.5
gcloud run deploy --image evylang/learn:v0.0.5 --region australia-southeast1 --allow-unauthenticated learnapi
```

something, something run <-> firestore integration

Authenticated enduser access only:
https://cloud.google.com/run/docs/authenticating/end-users
