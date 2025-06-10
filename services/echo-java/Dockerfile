FROM maven:3.9-eclipse-temurin-11-alpine AS build
WORKDIR /app
COPY pom.xml .
COPY src ./src
# Package the application, the JAR will be executable
RUN mvn package -DskipTests

FROM eclipse-temurin:11-jre-alpine
WORKDIR /app
# Copy the executable JAR from the build stage
COPY --from=build /app/target/echo-java-0.0.1-SNAPSHOT.jar echo-java.jar
EXPOSE 8083
CMD ["java", "-jar", "echo-java.jar"]
