FROM ubuntu
COPY dist/climkit-to-mqtt-amd64 /climkit-to-mqtt-amd64
COPY config.yaml.example /
RUN ls /
ENTRYPOINT ["/climkit-to-mqtt-amd64"]

