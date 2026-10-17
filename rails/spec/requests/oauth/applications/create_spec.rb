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

  it "退会済みのユーザーがアプリケーションを作成できないこと" do
    user = FactoryBot.create(:registered_user)
    login_as(user, scope: :user)
    user.update_column(:deleted_at, Time.current)

    expect {
      post "/oauth/applications", params: {
        oauth_application: {
          name: "Test App",
          redirect_uri: "http://example.com/callback",
          scopes: "read"
        }
      }
    }.not_to change(Oauth::Application, :count)

    expect(response).to have_http_status(:forbidden)
  end

  describe "退会受付と同時にアプリケーションを作成するとき" do
    self.use_transactional_tests = false

    it "ユーザー行のロック解除後に削除状態を読み直し、作成を拒否すること" do
      # セッションストアの行はユーザーと結び付かないため、テスト中に作られた行をidの範囲で消す。
      last_session_id = ActiveRecord::SessionStore::Session.maximum(:id) || 0
      locked = Queue.new
      release = Queue.new

      begin
        user = FactoryBot.create(:registered_user)
        login_as(user, scope: :user)
        get "/oauth/applications"
        expect(response).to have_http_status(:ok)

        blocker = Thread.new do
          ActiveRecord::Base.connection_pool.with_connection do |connection|
            connection.transaction do
              connection.exec_query("SELECT id FROM users WHERE id = #{user.id} FOR UPDATE")
              locked << connection.select_value("SELECT pg_backend_pid()")
              release.pop
              connection.exec_update("UPDATE users SET deleted_at = NOW() WHERE id = #{user.id}", "退会済みへ変更")
            end
          end
        end

        blocker_pid = locked.pop
        request_pid = ActiveRecord::Base.connection.select_value("SELECT pg_backend_pid()")
        waiting = Thread.new do
          ActiveRecord::Base.connection_pool.with_connection do |connection|
            Timeout.timeout(5, Timeout::Error, "アプリケーションの作成がユーザー行のロックを待機しなかった") do
              loop do
                blocked = connection.select_value(<<~SQL)
                  SELECT EXISTS (
                    SELECT 1 FROM pg_stat_activity
                    WHERE pid = #{request_pid}
                      AND #{blocker_pid} = ANY(pg_blocking_pids(pid))
                  )
                SQL
                break if blocked

                sleep 0.01
              end
            end
          end
        ensure
          release << true
        end

        post "/oauth/applications", params: {
          oauth_application: {
            name: "Test App",
            redirect_uri: "http://example.com/callback",
            scopes: "read"
          }
        }

        waiting.value
        blocker.value
        expect(response).to have_http_status(:forbidden)
        expect(Oauth::Application.where(owner: user)).not_to exist
      ensure
        release << true
        blocker&.join
        waiting&.join
        if user
          Oauth::Application.where(owner: user).delete_all
          user.destroy!
        end
        ActiveRecord::SessionStore::Session.where("id > ?", last_session_id).delete_all
      end
    end
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
