FROM alpine:3.6

ADD http://search.maven.org/remotecontent?filepath=org/jolokia/jolokia-jvm/1.4.0/jolokia-jvm-1.4.0-agent.jar /jolokia.jar
ADD https://repo1.maven.org/maven2/io/prometheus/jmx/jmx_prometheus_javaagent/0.3.0/jmx_prometheus_javaagent-0.3.0.jar /jmx_prometheus_javaagent.jar
ADD https://raw.githubusercontent.com/prometheus/jmx_exporter/0b490c14b6a8b53518b63aaaf02bf769e2eada4e/example_configs/cassandra.yml /jmx_prometheus_javaagent.yaml

RUN chmod a+r /jolokia.jar && touch /jolokia.jar
RUN chmod a+r /jmx_prometheus_javaagent.jar && touch /jmx_prometheus_javaagent.jar
RUN chmod a+r /jmx_prometheus_javaagent.yaml && touch /jmx_prometheus_javaagent.yaml

ADD https://github.com/jetstack/cassandra-kubernetes-seed-provider/releases/download/0.1.0/libcassandra-kubernetes-seed-provider.jar /libcassandra-kubernetes-seed-provider.jar
RUN chmod a+r /libcassandra-kubernetes-seed-provider.jar && touch /libcassandra-kubernetes-seed-provider.jar

ADD navigator-pilot-cassandra_linux_amd64 /pilot

ENTRYPOINT ["/pilot"]
