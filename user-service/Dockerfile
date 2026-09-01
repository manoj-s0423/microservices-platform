FROM eclipse-temurin:21-jdk-jammy AS build

WORKDIR /app

COPY pom.xml .

RUN mvn dependency:go-offline -B

COPY src ./src

RUN mvn package -DskipTests -B

FROM eclipse-temurin:21-jdk-jammy

WORKDIR /app

COPY --from=build /app/target/user-service.jar ./user-service.jar

RUN groupadd -r app && useradd -r -g app --no-create-home app

USER app

EXPOSE 8081

CMD ["java", "-jar", "user-service.jar"]