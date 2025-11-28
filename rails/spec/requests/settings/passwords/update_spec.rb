# typed: false
# frozen_string_literal: true

RSpec.describe "PATCH /settings/password", type: :request do
  it "ログイン済みユーザーが有効なパスワードでパスワードを更新できること" do
    user = create(:registered_user, password: "current_password")
    login_as(user, scope: :user)

    patch "/settings/password", params: {
      user: {
        current_password: "current_password",
        password: "new_password123",
        password_confirmation: "new_password123"
      }
    }

    expect(response).to redirect_to(settings_password_path)
    expect(flash[:notice]).to eq(I18n.t("messages._common.updated"))
    expect(user.reload.valid_password?("new_password123")).to be true
  end

  it "現在のパスワードが間違っている場合、更新が失敗すること" do
    user = create(:registered_user, password: "current_password")
    login_as(user, scope: :user)

    patch "/settings/password", params: {
      user: {
        current_password: "wrong_password",
        password: "new_password123",
        password_confirmation: "new_password123"
      }
    }

    expect(response.status).to eq(422)
    expect(user.reload.valid_password?("current_password")).to be true
    expect(user.valid_password?("new_password123")).to be false
  end

  it "新しいパスワードと確認パスワードが一致しない場合、更新が失敗すること" do
    user = create(:registered_user, password: "current_password")
    login_as(user, scope: :user)

    patch "/settings/password", params: {
      user: {
        current_password: "current_password",
        password: "new_password123",
        password_confirmation: "different_password"
      }
    }

    expect(response.status).to eq(422)
    expect(user.reload.valid_password?("current_password")).to be true
    expect(user.valid_password?("new_password123")).to be false
  end

  it "新しいパスワードが短すぎる場合、更新が失敗すること" do
    user = create(:registered_user, password: "current_password")
    login_as(user, scope: :user)

    patch "/settings/password", params: {
      user: {
        current_password: "current_password",
        password: "123",
        password_confirmation: "123"
      }
    }

    expect(response.status).to eq(422)
    expect(user.reload.valid_password?("current_password")).to be true
    expect(user.valid_password?("123")).to be false
  end

  it "新しいパスワードが空の場合、更新が失敗すること" do
    user = create(:registered_user, password: "current_password")
    login_as(user, scope: :user)

    patch "/settings/password", params: {
      user: {
        current_password: "current_password",
        password: "",
        password_confirmation: ""
      }
    }

    expect(response.status).to eq(422)
    expect(user.reload.valid_password?("current_password")).to be true
  end

  it "現在のパスワードが空の場合、更新が失敗すること" do
    user = create(:registered_user, password: "current_password")
    login_as(user, scope: :user)

    patch "/settings/password", params: {
      user: {
        current_password: "",
        password: "new_password123",
        password_confirmation: "new_password123"
      }
    }

    expect(response.status).to eq(422)
    expect(user.reload.valid_password?("current_password")).to be true
    expect(user.valid_password?("new_password123")).to be false
  end

  it "未ログインユーザーはログインページにリダイレクトされること" do
    patch "/settings/password", params: {
      user: {
        current_password: "current_password",
        password: "new_password123",
        password_confirmation: "new_password123"
      }
    }

    expect(response).to redirect_to(new_user_session_path)
  end
end
