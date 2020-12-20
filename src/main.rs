use actix_web::App;
use actix_web::HttpResponse;
use actix_web::HttpServer;
use actix_web::middleware;
use actix_web::web;
use serde::Deserialize;
use serde::Serialize;

#[actix_web::main]
async fn main() -> std::io::Result<()> {
    std::env::set_var("RUST_LOG", "actix_web=debug");
    env_logger::init();

    HttpServer::new(|| {
        App::new()
            .wrap(middleware::Logger::default())
            .data(web::JsonConfig::default().limit(4096))
            .service(web::resource("/").route(web::post().to(index)))
    })
    .bind("127.0.0.1:8080")?
    .run()
    .await
}

#[derive(Debug, Serialize, Deserialize)]
struct SlackEvent {
    challenge: Option<String>,
}

#[derive(Debug, Serialize, Deserialize)]
struct ChallengeResponse {
    challenge: String,
}

async fn index(item: web::Json<SlackEvent>) -> HttpResponse {
    println!("item: {:?}", &item);
    println!("item.0: {:?}", item.0);
    match item.0.challenge {
        Some(challenge) => {
            HttpResponse::Ok().json(ChallengeResponse { challenge })
        },
        _ => {
            HttpResponse::Ok().json("ok")
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use actix_web::dev::Service;
    use actix_web::{http, test, web, App, Error};

    #[actix_rt::test]
    async fn test_index_challenge() -> Result<(), Error> {
        let mut app = test::init_service(
            App::new().service(web::resource("/").route(web::post().to(index))),
        )
        .await;

        let req = test::TestRequest::post()
            .uri("/")
            .set_json(&SlackEvent {
                challenge: Some("challengetoken".to_owned()),
            })
            .to_request();
        let resp = app.call(req).await.unwrap();

        assert_eq!(resp.status(), http::StatusCode::OK);

        let response_body = match resp.response().body().as_ref() {
            Some(actix_web::body::Body::Bytes(bytes)) => bytes,
            _ => panic!("Response error"),
        };

        assert_eq!(response_body, r##"{"challenge":"challengetoken"}"##);

        Ok(())
    }
}
