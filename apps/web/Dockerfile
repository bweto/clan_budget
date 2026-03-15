FROM node:24-alpine AS base

WORKDIR /app

# In development, we mount the volume, so we just need a command to keep it running
# or to start next dev directly
CMD ["npm", "run", "dev"]
