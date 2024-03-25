FROM alpine:3.14

WORKDIR /app

# we expect the user to provide the binary path available at the build context
ARG SESSION_SECRET
ARG CORS_ALLOWED_ORIGINS
ARG DOMAIN

COPY ./kgw.base-config.json ./config/config.json

# the command needed is
# /app/kgw --config ./config/config.json --session-secret $SESSION_SECRET \
#    --cors-allow-origins $CORS_ALLOWED_ORIGINS --domain $DOMAIN"

ENV SESSION_SECRET=$SESSION_SECRET
ENV CORS_ALLOWED_ORIGINS=$CORS_ALLOWED_ORIGINS
ENV DOMAIN=$DOMAIN
# cors args should only be set if the user has provided a value, otherwise it will be empty
ENV CORS_ARGS=""
RUN if [ -n "$CORS_ALLOWED_ORIGINS" ]; then \
    CORS_ARGS="--cors-allow-origins $CORS_ALLOWED_ORIGINS"; \
    fi

CMD /app/kgw --config ./config/config.json --session-secret $SESSION_SECRET \
     --domain $DOMAIN $CORS_ARGS