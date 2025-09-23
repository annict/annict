# typed: false
# frozen_string_literal: true

RSpec.describe "GET /settings/email_notification/unsubscribe", type: :request do
  it "有効なkeyとaction_nameパラメータで適切なイベント通知が無効化されること" do
    user = FactoryBot.create(:registered_user)
    email_notification = user.email_notification
    email_notification.update!(
      unsubscription_key: "test-key-123",
      event_followed_user: true
    )

    get "/settings/email_notification/unsubscribe", params: {
      key: "test-key-123",
      action_name: "followed_user"
    }

    expect(response.status).to eq(200)
    email_notification.reload
    expect(email_notification.event_followed_user).to eq(false)
  end

  it "keyパラメータが空の場合、ルートページにリダイレクトされること" do
    get "/settings/email_notification/unsubscribe", params: {
      key: "",
      action_name: "followed_user"
    }

    expect(response).to redirect_to(root_path)
  end

  it "action_nameパラメータが空の場合、ルートページにリダイレクトされること" do
    user = FactoryBot.create(:registered_user)
    email_notification = user.email_notification
    email_notification.update!(unsubscription_key: "test-key-123")

    get "/settings/email_notification/unsubscribe", params: {
      key: "test-key-123",
      action_name: ""
    }

    expect(response).to redirect_to(root_path)
  end

  it "keyパラメータが存在しない場合、ルートページにリダイレクトされること" do
    get "/settings/email_notification/unsubscribe", params: {
      action_name: "followed_user"
    }

    expect(response).to redirect_to(root_path)
  end

  it "action_nameパラメータが存在しない場合、ルートページにリダイレクトされること" do
    get "/settings/email_notification/unsubscribe", params: {
      key: "test-key-123"
    }

    expect(response).to redirect_to(root_path)
  end

  it "無効なunsubscription_keyの場合、404エラーが返されること" do
    get "/settings/email_notification/unsubscribe", params: {
      key: "invalid-key",
      action_name: "followed_user"
    }

    expect(response).to have_http_status(:not_found)
  end

  it "存在しないイベントカラムの場合、RoutingErrorが発生すること" do
    user = FactoryBot.create(:registered_user)
    email_notification = user.email_notification
    email_notification.update!(unsubscription_key: "test-key-123")

    get "/settings/email_notification/unsubscribe", params: {
      key: "test-key-123",
      action_name: "invalid_event"
    }

    expect(response).to have_http_status(:not_found)
  end

  it "liked_episode_recordイベントの通知が無効化されること" do
    user = FactoryBot.create(:registered_user)
    email_notification = user.email_notification
    email_notification.update!(
      unsubscription_key: "test-key-456",
      event_liked_episode_record: true
    )

    get "/settings/email_notification/unsubscribe", params: {
      key: "test-key-456",
      action_name: "liked_episode_record"
    }

    expect(response.status).to eq(200)
    email_notification.reload
    expect(email_notification.event_liked_episode_record).to eq(false)
  end

  it "favorite_works_addedイベントの通知が無効化されること" do
    user = FactoryBot.create(:registered_user)
    email_notification = user.email_notification
    email_notification.update!(
      unsubscription_key: "test-key-789",
      event_favorite_works_added: true
    )

    get "/settings/email_notification/unsubscribe", params: {
      key: "test-key-789",
      action_name: "favorite_works_added"
    }

    expect(response.status).to eq(200)
    email_notification.reload
    expect(email_notification.event_favorite_works_added).to eq(false)
  end

  it "related_works_addedイベントの通知が無効化されること" do
    user = FactoryBot.create(:registered_user)
    email_notification = user.email_notification
    email_notification.update!(
      unsubscription_key: "test-key-abc",
      event_related_works_added: true
    )

    get "/settings/email_notification/unsubscribe", params: {
      key: "test-key-abc",
      action_name: "related_works_added"
    }

    expect(response.status).to eq(200)
    email_notification.reload
    expect(email_notification.event_related_works_added).to eq(false)
  end

  it "すでにfalseになっているイベントでも正常に処理されること" do
    user = FactoryBot.create(:registered_user)
    email_notification = user.email_notification
    email_notification.update!(
      unsubscription_key: "test-key-def",
      event_followed_user: false
    )

    get "/settings/email_notification/unsubscribe", params: {
      key: "test-key-def",
      action_name: "followed_user"
    }

    expect(response.status).to eq(200)
    email_notification.reload
    expect(email_notification.event_followed_user).to eq(false)
  end
end
