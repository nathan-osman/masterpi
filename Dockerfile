FROM golang:latest


FROM scratch
MAINTAINER Nathan Osman <nathan@quickmediasolutions.com>

# Add the executable
ADD dist/linux_arm/masterpi /usr/local/bin/

# Copy the timezone data from the golang:latest image
COPY --from=0 /usr/share/zoneinfo /usr/share/zoneinfo

# Set the entrypoint for the container
ENTRYPOINT ["/usr/local/bin/masterpi"]
