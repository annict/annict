# typed: false
# frozen_string_literal: true

RSpec.describe "GET|POST /users/auth/gumroad", type: :request do
  it "ログインしていない場合、Gumroadの認証を開始すること" do
    post "/users/auth/gumroad"

    expect(response).to have_http_status(:found)
  end

  it "ログインしている場合、Gumroadの認証を開始すること" do
    user = FactoryBot.create(:registered_user)
    login_as(user, scope: :user)

    post "/users/auth/gumroad"

    expect(response).to have_http_status(:found)
  end
end

RSpec.describe "GET|POST /users/auth/gumroad/callback", type: :request do
  it "ログインしていない場合で既存のプロバイダーが存在する場合、ユーザーにログインすること" do
    auth_hash = {
      provider: "gumroad",
      uid: "gumroad123",
      info: {
        email: "test@example.com"
      },
      credentials: {
        token: "gumroad_token",
        expires_at: 1234567890
      }
    }

    OmniAuth.config.test_mode = true
    OmniAuth.config.add_mock(:gumroad, auth_hash)

    # Gumroad APIの呼び出しをモック
    allow_any_instance_of(GumroadClient).to receive(:fetch_subscriber_by_email).and_return(nil)

    user = FactoryBot.create(:registered_user)
    FactoryBot.create(:provider, name: "gumroad", uid: "gumroad123", user: user)

    get "/users/auth/gumroad/callback"

    expect(response).to have_http_status(:found)
    expect(response).to redirect_to(root_path)
    expect(controller.current_user).to eq(user)

    OmniAuth.config.test_mode = false
    OmniAuth.config.mock_auth[:gumroad] = nil
  end

  it "ログインしていない場合で既存のプロバイダーが存在しない場合、エラーメッセージを表示すること" do
    auth_hash = {
      provider: "gumroad",
      uid: "gumroad123",
      info: {
        email: "test@example.com"
      },
      credentials: {
        token: "gumroad_token",
        expires_at: 1234567890
      }
    }

    OmniAuth.config.test_mode = true
    OmniAuth.config.add_mock(:gumroad, auth_hash)

    # Gumroad APIの呼び出しをモック
    allow_any_instance_of(GumroadClient).to receive(:fetch_subscriber_by_email).and_return(nil)

    get "/users/auth/gumroad/callback"

    expect(response).to have_http_status(:found)
    expect(response).to redirect_to(root_path)
    expect(request.flash[:alert]).to eq(I18n.t("messages.callbacks.sign_in_failed"))

    OmniAuth.config.test_mode = false
    OmniAuth.config.mock_auth[:gumroad] = nil
  end

  it "ログインしている場合でGumroadプロバイダーの場合、Gumroadのサブスクライバーが見つからない場合はエラーメッセージを表示すること" do
    auth_hash = {
      provider: "gumroad",
      uid: "gumroad123",
      info: {
        email: "test@example.com"
      },
      credentials: {
        token: "gumroad_token",
        expires_at: 1234567890
      }
    }

    OmniAuth.config.test_mode = true
    OmniAuth.config.add_mock(:gumroad, auth_hash)

    # Gumroad APIの呼び出しをモック
    allow_any_instance_of(GumroadClient).to receive(:fetch_subscriber_by_email).and_return(nil)

    user = FactoryBot.create(:registered_user)
    login_as(user, scope: :user)

    allow_any_instance_of(Forms::SupporterRegistrationForm).to receive(:subscriber).and_return(nil)

    get "/users/auth/gumroad/callback"

    expect(response).to have_http_status(:found)
    expect(response).to redirect_to(supporters_path)
    expect(request.flash[:alert]).to eq(I18n.t("messages.supporters.gumroad_subscriber_not_found"))

    OmniAuth.config.test_mode = false
    OmniAuth.config.mock_auth[:gumroad] = nil
  end

  it "ログインしている場合でGumroadプロバイダーの場合、フォームが無効な場合はエラーメッセージを表示すること" do
    auth_hash = {
      provider: "gumroad",
      uid: "gumroad123",
      info: {
        email: "test@example.com"
      },
      credentials: {
        token: "gumroad_token",
        expires_at: 1234567890
      }
    }

    OmniAuth.config.test_mode = true
    OmniAuth.config.add_mock(:gumroad, auth_hash)

    # Gumroad APIの呼び出しをモック
    allow_any_instance_of(GumroadClient).to receive(:fetch_subscriber_by_email).and_return(nil)

    user = FactoryBot.create(:registered_user)
    login_as(user, scope: :user)

    gumroad_subscriber = FactoryBot.create(:gumroad_subscriber)

    allow_any_instance_of(Forms::SupporterRegistrationForm).to receive(:subscriber).and_return(gumroad_subscriber)
    allow_any_instance_of(Forms::SupporterRegistrationForm).to receive(:invalid?).and_return(true)
    allow_any_instance_of(Forms::SupporterRegistrationForm).to receive(:errors).and_return(
      instance_double("errors", full_messages: ["エラーメッセージ"])
    )

    get "/users/auth/gumroad/callback"

    expect(response).to have_http_status(:found)
    expect(response).to redirect_to(supporters_path)
    expect(request.flash[:alert]).to eq("エラーメッセージ")

    OmniAuth.config.test_mode = false
    OmniAuth.config.mock_auth[:gumroad] = nil
  end

  it "ログインしている場合でGumroadプロバイダーの場合、正常な場合はサポーター登録を実行すること" do
    auth_hash = {
      provider: "gumroad",
      uid: "gumroad123",
      info: {
        email: "test@example.com"
      },
      credentials: {
        token: "gumroad_token",
        expires_at: 1234567890
      }
    }

    OmniAuth.config.test_mode = true
    OmniAuth.config.add_mock(:gumroad, auth_hash)

    # Gumroad APIの呼び出しをモック
    allow_any_instance_of(GumroadClient).to receive(:fetch_subscriber_by_email).and_return(nil)

    user = FactoryBot.create(:registered_user)
    login_as(user, scope: :user)

    gumroad_subscriber = FactoryBot.create(:gumroad_subscriber)
    creator_mock = instance_double("SupporterRegistrationCreator")

    allow_any_instance_of(Forms::SupporterRegistrationForm).to receive(:subscriber).and_return(gumroad_subscriber)
    allow_any_instance_of(Forms::SupporterRegistrationForm).to receive(:invalid?).and_return(false)
    allow(Creators::SupporterRegistrationCreator).to receive(:new).and_return(creator_mock)
    allow(creator_mock).to receive(:call)

    get "/users/auth/gumroad/callback"

    expect(response).to have_http_status(:found)
    expect(response).to redirect_to(root_path)
    expect(request.flash[:notice]).to eq(I18n.t("messages._common.connected"))
    expect(Creators::SupporterRegistrationCreator).to have_received(:new).with(
      user: user,
      form: instance_of(Forms::SupporterRegistrationForm)
    )
    expect(creator_mock).to have_received(:call)

    OmniAuth.config.test_mode = false
    OmniAuth.config.mock_auth[:gumroad] = nil
  end

  it "POST /users/auth/gumroad/callback でログインしている場合でGumroadプロバイダーの場合、正常な場合はサポーター登録を実行すること" do
    auth_hash = {
      provider: "gumroad",
      uid: "gumroad123",
      info: {
        email: "test@example.com"
      },
      credentials: {
        token: "gumroad_token",
        expires_at: 1234567890
      }
    }

    OmniAuth.config.test_mode = true
    OmniAuth.config.add_mock(:gumroad, auth_hash)

    # Gumroad APIの呼び出しをモック
    allow_any_instance_of(GumroadClient).to receive(:fetch_subscriber_by_email).and_return(nil)

    user = FactoryBot.create(:registered_user)
    login_as(user, scope: :user)

    gumroad_subscriber = FactoryBot.create(:gumroad_subscriber)
    creator_mock = instance_double("SupporterRegistrationCreator")

    allow_any_instance_of(Forms::SupporterRegistrationForm).to receive(:subscriber).and_return(gumroad_subscriber)
    allow_any_instance_of(Forms::SupporterRegistrationForm).to receive(:invalid?).and_return(false)
    allow(Creators::SupporterRegistrationCreator).to receive(:new).and_return(creator_mock)
    allow(creator_mock).to receive(:call)

    post "/users/auth/gumroad/callback"

    expect(response).to have_http_status(:found)
    expect(response).to redirect_to(root_path)
    expect(request.flash[:notice]).to eq(I18n.t("messages._common.connected"))
    expect(Creators::SupporterRegistrationCreator).to have_received(:new).with(
      user: user,
      form: instance_of(Forms::SupporterRegistrationForm)
    )
    expect(creator_mock).to have_received(:call)

    OmniAuth.config.test_mode = false
    OmniAuth.config.mock_auth[:gumroad] = nil
  end
end
