use std::env;
use std::fs::File;
use std::io::{BufReader, Read};
use std::net::{IpAddr, Ipv4Addr};

use lazy_static::lazy_static;

use hyper::header::{HeaderValue, CONTENT_TYPE};

use qrcode_generator::QrCodeEcc;

use warp::http::Uri;
use warp::reply::Response;
use warp::Filter;
use warp::Reply;

#[tokio::main]
async fn main() {
    let log = warp::log::custom(|info: warp::filters::log::Info| {
        println!("{} {} {}", info.method(), info.path(), info.status());
    });

    let home = warp::path::end().map(serve_homepage);
    let qr_code_endpoint = warp::path("qrcode").map(serve_qr_image);
    let qr_endpoint = warp::path!("qr").map(redirect_to_rick_astley);

    let routes = home.or(qr_code_endpoint).or(qr_endpoint).with(log);

    let (ip, port) = server_address();
    println!("Listening on {:?}:{}...", ip, port);
    warp::serve(routes).run((ip, port)).await;
}

fn server_address() -> (IpAddr, u16) {
    let address = env::var("SERVER_ADDRESS")
        .unwrap_or_else(|_| "".to_string())
        .parse::<IpAddr>()
        .unwrap_or_else(|_| IpAddr::V4(Ipv4Addr::new(127, 0, 0, 1)));

    let port = env::var("SERVER_PORT")
        .unwrap_or_else(|_| "".to_string())
        .parse::<u16>()
        .unwrap_or(8080);

    (address, port)
}

fn serve_homepage() -> impl warp::Reply {
    let mut contents = String::new();
    let mut reader = BufReader::new(File::open("static/index.html").unwrap());
    reader.read_to_string(&mut contents).unwrap();

    return warp::reply::html(contents);
}

lazy_static! {
    static ref QRCODE_DATA: Vec<u8> =
        qrcode_generator::to_png_to_vec("https://crockeo.net/qr", QrCodeEcc::Low, 1024).unwrap();
}

struct QrCode {}

impl Reply for QrCode {
    fn into_response(self) -> Response {
        let qrcode_data_slice = QRCODE_DATA.as_slice();
        let mut res = Response::new(qrcode_data_slice.into());
        res.headers_mut()
            .insert(CONTENT_TYPE, HeaderValue::from_static("image/png"));
        res
    }
}

fn serve_qr_image() -> impl warp::Reply {
    QrCode {}
}

fn redirect_to_rick_astley() -> impl warp::Reply {
    warp::redirect::temporary(Uri::from_static(
        "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
    ))
}
