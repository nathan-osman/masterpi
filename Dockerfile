FROM scratch
MAINTAINER Nathan Osman <nathan@quickmediasolutions.com>

# Add the executable
ADD dist/linux_arm/masterpi /usr/local/bin/

# Set the entrypoint for the container
ENTRYPOINT ["/usr/local/bin/masterpi"]
