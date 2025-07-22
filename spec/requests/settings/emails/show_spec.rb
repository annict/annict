# typed: false
# frozen_string_literal: true

RSpec.describe "GET /settings/email", type: :request do
  it "ログイン済みユーザーはページが表示されること" do
    user = create(:registered_user)
    login_as(user, scope: :user)

    get "/settings/email"

    expect(response.status).to eq(200)
    expect(response.body).to include("メールアドレス")
  end

  it "未ログインユーザーはログインページにリダイレクトされること" do
    get "/settings/email"

    expect(response).to redirect_to(new_user_session_path)
  end

  it "ユーザーの現在のメールアドレスが表示されること" do
    user = create(:registered_user, email: "test@example.com")
    login_as(user, scope: :user)

    get "/settings/email"

    expect(response.status).to eq(200)
    expect(response.body).to include("test@example.com")
  end
end

RSpec.describe "PATCH /settings/email", type: :request do
  it "ログイン済みユーザーが有効なメールアドレスで更新するとリダイレクトされること" do
    user = create(:registered_user)
    login_as(user, scope: :user)

    patch "/settings/email", params: {
      forms_user_email_form: {
        email: "new@example.com"
      }
    }

    expect(response).to redirect_to(settings_email_path)
    expect(flash[:notice]).to eq(I18n.t("messages.accounts.email_sent_for_confirmation"))
  end

  it "未ログインユーザーはログインページにリダイレクトされること" do
    patch "/settings/email", params: {
      forms_user_email_form: {
        email: "new@example.com"
      }
    }

    expect(response).to redirect_to(new_user_session_path)
  end

  it "無効なメールアドレスの場合、422が返されること" do
    user = create(:registered_user)
    login_as(user, scope: :user)

    patch "/settings/email", params: {
      forms_user_email_form: {
        email: "invalid"
      }
    }

    expect(response.status).to eq(422)
    expect(response.body).to include("メールアドレス")
  end

  it "空のメールアドレスの場合、422が返されること" do
    user = create(:registered_user)
    login_as(user, scope: :user)

    patch "/settings/email", params: {
      forms_user_email_form: {
        email: ""
      }
    }

    expect(response.status).to eq(422)
    expect(response.body).to include("メールアドレス")
  end

  it "既に使用されているメールアドレスの場合、422が返されること" do
    user = create(:registered_user)
    create(:registered_user, email: "taken@example.com")
    login_as(user, scope: :user)

    patch "/settings/email", params: {
      forms_user_email_form: {
        email: "taken@example.com"
      }
    }

    expect(response.status).to eq(422)
    expect(response.body).to include("メールアドレス")
  end

  it "メールアドレスの前後の空白が削除されること" do
    user = create(:registered_user)
    login_as(user, scope: :user)

    patch "/settings/email", params: {
      forms_user_email_form: {
        email: "  new@example.com  "
      }
    }

    expect(response).to redirect_to(settings_email_path)
    # confirm_to_update_email!メソッドが呼ばれているか確認
    email_confirmation = EmailConfirmation.last
    expect(email_confirmation.email).to eq("new@example.com")
  end
end
