# typed: false
# frozen_string_literal: true

RSpec.describe "POST /legacy/sign_in", type: :request do
  it "確認済みのユーザーがログインできること" do
    user = create(:registered_user)
    password = "password"
    user.update!(password: password)

    post "/legacy/sign_in", params: {
      user: {
        email_username: user.username,
        password: password
      }
    }

    expect(response).to redirect_to("/")
    expect(flash[:notice]).to include("ログインしました")
  end

  it "2020年4月7日以降に登録した未確認ユーザーがログインできないこと" do
    user = create(:registered_user, created_at: Time.zone.parse("2020-04-07 00:00:01"))
    password = "password"
    user.update!(password: password, confirmed_at: nil)

    post "/legacy/sign_in", params: {
      user: {
        email_username: user.email,
        password: password
      }
    }

    expect(response).to redirect_to("/")
    expect(flash[:alert]).to include("メールアドレスの確認が完了していないため、ログインできませんでした")
  end

  it "存在しないユーザーでログインできないこと" do
    post "/legacy/sign_in", params: {
      user: {
        email_username: "nonexistent@example.com",
        password: "password"
      }
    }

    expect(response).to have_http_status(:ok)
    expect(flash[:alert]).to include("ユーザが見つかりません")
  end

  it "パスワードが間違っているときログインできないこと" do
    user = create(:registered_user)
    password = "password"
    user.update!(password: password)

    post "/legacy/sign_in", params: {
      user: {
        email_username: user.username,
        password: "wrong_password"
      }
    }

    expect(response).to have_http_status(:ok)
    expect(flash[:alert]).to include("ログインできません")
  end
end
