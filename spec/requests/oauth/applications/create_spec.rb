# typed: false
# frozen_string_literal: true

RSpec.describe "POST /oauth/applications", type: :request do
  it "ログインしていないとき、ログインページにリダイレクトされること" do
    post "/oauth/applications", params: {
      oauth_application: {
        name: "Test App",
        redirect_uri: "http://example.com/callback",
        scopes: "read write"
      }
    }

    expect(response.status).to eq(302)
    expect(response.location).to start_with("http://api.annict.test:3000/sign_in")
  end

  it "管理者でないユーザーがアプリケーションを作成したとき、アプリケーションが作成されること" do
    user = FactoryBot.create(:registered_user)
    login_as(user, scope: :user)

    expect {
      post "/oauth/applications", params: {
        oauth_application: {
          name: "Test App",
          redirect_uri: "http://example.com/callback",
          scopes: "read write"
        }
      }
    }.to change(Oauth::Application, :count).by(1)

    expect(response.status).to eq(302)
    created_app = Oauth::Application.last
    expect(response.location).to eq("http://api.annict.test:3000/oauth/applications/#{created_app.id}")
    expect(created_app.name).to eq("Test App")
    expect(created_app.redirect_uri).to eq("http://example.com/callback")
    expect(created_app.scopes.to_s).to eq("read write")
    expect(created_app.owner).to eq(user)
  end

  it "管理者ユーザーがアプリケーションを作成したとき、アプリケーションが作成されること" do
    user = FactoryBot.create(:registered_user, :with_admin_role)
    login_as(user, scope: :user)

    expect {
      post "/oauth/applications", params: {
        oauth_application: {
          name: "Admin App",
          redirect_uri: "https://admin.example.com/callback",
          scopes: "read"
        }
      }
    }.to change(Oauth::Application, :count).by(1)

    expect(response.status).to eq(302)
    created_app = Oauth::Application.last
    expect(response.location).to eq("http://api.annict.test:3000/oauth/applications/#{created_app.id}")
    expect(created_app.name).to eq("Admin App")
    expect(created_app.redirect_uri).to eq("https://admin.example.com/callback")
    expect(created_app.scopes.to_s).to eq("read")
    expect(created_app.owner).to eq(user)
  end

  it "必須パラメータが不足しているとき、アプリケーションが作成されず、フォームが再表示されること" do
    user = FactoryBot.create(:registered_user)
    login_as(user, scope: :user)

    expect {
      post "/oauth/applications", params: {
        oauth_application: {
          name: "",
          redirect_uri: "",
          scopes: ""
        }
      }
    }.not_to change(Oauth::Application, :count)

    expect(response.status).to eq(200)
    expect(response.body).to include("form")
  end

  it "不正なリダイレクトURIが指定されたとき、アプリケーションが作成されず、フォームが再表示されること" do
    user = FactoryBot.create(:registered_user)
    login_as(user, scope: :user)

    expect {
      post "/oauth/applications", params: {
        oauth_application: {
          name: "Test App",
          redirect_uri: "invalid-uri",
          scopes: "read"
        }
      }
    }.not_to change(Oauth::Application, :count)

    expect(response.status).to eq(200)
    expect(response.body).to include("form")
  end

  it "複数のリダイレクトURIが改行区切りで指定されたとき、アプリケーションが作成されること" do
    user = FactoryBot.create(:registered_user)
    login_as(user, scope: :user)

    expect {
      post "/oauth/applications", params: {
        oauth_application: {
          name: "Multi Redirect App",
          redirect_uri: "http://example.com/callback1\nhttp://example.com/callback2",
          scopes: "read write"
        }
      }
    }.to change(Oauth::Application, :count).by(1)

    expect(response.status).to eq(302)
    created_app = Oauth::Application.last
    expect(created_app.redirect_uri).to eq("http://example.com/callback1\nhttp://example.com/callback2")
  end

  it "不正なスコープが指定されたとき、アプリケーションが作成されること" do
    user = FactoryBot.create(:registered_user)
    login_as(user, scope: :user)

    expect {
      post "/oauth/applications", params: {
        oauth_application: {
          name: "Test App",
          redirect_uri: "http://example.com/callback",
          scopes: "read write invalid_scope"
        }
      }
    }.to change(Oauth::Application, :count).by(1)

    expect(response.status).to eq(302)
    created_app = Oauth::Application.last
    expect(created_app.scopes.to_s).to eq("read write invalid_scope")
  end

  it "日本語の名前でアプリケーションを作成したとき、アプリケーションが作成されること" do
    user = FactoryBot.create(:registered_user)
    login_as(user, scope: :user)

    expect {
      post "/oauth/applications", params: {
        oauth_application: {
          name: "テストアプリケーション",
          redirect_uri: "http://example.com/callback",
          scopes: "read"
        }
      }
    }.to change(Oauth::Application, :count).by(1)

    expect(response.status).to eq(302)
    created_app = Oauth::Application.last
    expect(created_app.name).to eq("テストアプリケーション")
  end
end
