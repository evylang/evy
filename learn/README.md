```
docker buildx build --platform linux/amd64 -t evylang/learn:v0.0.5 .
docker push evylang/learn:v0.0.5
gcloud run deploy --image evylang/learn:v0.0.5 --region australia-southeast1 --allow-unauthenticated learnapi
```

local run:

```
docker run --rm -p 8080:8080 -e EVY_FIREBASE_CREDENTIAL_FILE=/data/evy-lang-test-firebase-adminsdk-5ud3e-4ef53c5971.json -v /Users/julia/Development/:/data/ evylang/learn:v0.0.9
```

something, something run <-> firestore integration

Authenticated enduser access only:
https://cloud.google.com/run/docs/authenticating/end-users
