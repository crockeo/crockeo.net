use std::fs::File;
use std::io::{BufReader, Read};

use lazy_static::lazy_static;

use hyper::header::{CONTENT_TYPE, HeaderValue};

use qrcode_generator::QrCodeEcc;

use warp::Filter;
use warp::Reply;
use warp::http::Uri;
use warp::reply::Response;

#[tokio::main]
async fn main() {
    let home = warp::path::end().map(serve_homepage);
    let qr_code_endpoint = warp::path("qrcode").map(serve_qr_image);
    let qr_endpoint = warp::path!("qr").map(redirect_to_rick_astley);

    let routes = home
	.or(qr_code_endpoint)
	.or(qr_endpoint);
    warp::serve(routes).run(([0, 0, 0, 0], 8080)).await;
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
	res.headers_mut().insert(CONTENT_TYPE, HeaderValue::from_static("image/png"));
	res
    }
}

fn serve_qr_image() -> impl warp::Reply {
    QrCode {}
}

fn redirect_to_rick_astley() -> impl warp::Reply {
    warp::redirect::temporary(Uri::from_static("https://www.youtube.com/watch?v=dQw4w9WgXcQ"))
}
