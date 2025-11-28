# typed: false
# frozen_string_literal: true

RSpec.describe "PATCH /settings/account", type: :request do
  it "ログイン済みユーザーがユーザー名を更新できること" do
    user = create(:registered_user)
    login_as(user, scope: :user)

    patch "/settings/account", params: {
      user: {
        username: "new_username"
      }
    }

    expect(response).to redirect_to("#{ENV["ANNICT_URL"]}/settings/account")
    expect(user.reload.username).to eq("new_username")
  end

  it "ログイン済みユーザーがタイムゾーンを更新できること" do
    user = create(:registered_user)
    login_as(user, scope: :user)

    patch "/settings/account", params: {
      user: {
        time_zone: "America/New_York"
      }
    }

    expect(response).to redirect_to("#{ENV["ANNICT_URL"]}/settings/account")
    expect(user.reload.time_zone).to eq("America/New_York")
  end

  it "ログイン済みユーザーがロケールを更新できること" do
    user = create(:registered_user)
    login_as(user, scope: :user)

    patch "/settings/account", params: {
      user: {
        locale: "en"
      }
    }

    expect(response).to redirect_to("#{ENV["ANNICT_EN_URL"]}/settings/account")
    expect(user.reload.locale).to eq("en")
  end

  it "ログイン済みユーザーが許可されたロケールを更新できること" do
    user = create(:registered_user)
    login_as(user, scope: :user)

    patch "/settings/account", params: {
      user: {
        allowed_locales: ["ja", "en"]
      }
    }

    expect(response).to redirect_to("#{ENV["ANNICT_URL"]}/settings/account")
    expect(user.reload.allowed_locales).to eq(["ja", "en"])
  end

  it "ログイン済みユーザーが複数の属性を同時に更新できること" do
    user = create(:registered_user)
    login_as(user, scope: :user)

    patch "/settings/account", params: {
      user: {
        username: "updated_user",
        time_zone: "Asia/Tokyo",
        locale: "ja",
        allowed_locales: ["ja"]
      }
    }

    expect(response).to redirect_to("#{ENV["ANNICT_URL"]}/settings/account")
    user.reload
    expect(user.username).to eq("updated_user")
    expect(user.time_zone).to eq("Asia/Tokyo")
    expect(user.locale).to eq("ja")
    expect(user.allowed_locales).to eq(["ja"])
  end

  it "無効なユーザー名形式の場合、更新が失敗すること" do
    user = create(:registered_user)
    login_as(user, scope: :user)

    patch "/settings/account", params: {
      user: {
        username: "invalid-username!"
      }
    }

    expect(response.status).to eq(422)
    expect(response.body).to include("基本情報")
    expect(user.reload.username).not_to eq("invalid-username!")
  end

  it "21文字以上のユーザー名の場合、更新が失敗すること" do
    user = create(:registered_user)
    login_as(user, scope: :user)

    patch "/settings/account", params: {
      user: {
        username: "a" * 21
      }
    }

    expect(response.status).to eq(422)
    expect(response.body).to include("基本情報")
    expect(user.reload.username).not_to eq("a" * 21)
  end

  it "既に使用されているユーザー名の場合、更新が失敗すること" do
    create(:registered_user, username: "existing_user")
    user = create(:registered_user)
    login_as(user, scope: :user)

    patch "/settings/account", params: {
      user: {
        username: "existing_user"
      }
    }

    expect(response.status).to eq(422)
    expect(response.body).to include("基本情報")
    expect(user.reload.username).not_to eq("existing_user")
  end

  it "無効なロケール値の場合、更新が失敗すること" do
    user = create(:registered_user)
    login_as(user, scope: :user)

    patch "/settings/account", params: {
      user: {
        locale: "invalid"
      }
    }

    expect(response.status).to eq(422)
    expect(response.body).to include("基本情報")
    expect(user.reload.locale).not_to eq("invalid")
  end

  it "未ログインユーザーはログインページにリダイレクトされること" do
    patch "/settings/account", params: {
      user: {
        username: "new_username"
      }
    }

    expect(response).to redirect_to(new_user_session_path)
  end
end
