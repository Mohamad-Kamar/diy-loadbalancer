FROM maven:3.9-eclipse-temurin-21-alpine AS build
WORKDIR /app
COPY pom.xml .
COPY src ./src
RUN mvn package -DskipTests

FROM eclipse-temurin:21-jre-alpine
WORKDIR /app
COPY --from=build /app/target/echo-java-0.0.1-SNAPSHOT.jar .
EXPOSE 8083
CMD ["java", "-jar", "echo-java-0.0.1-SNAPSHOT.jar"]
