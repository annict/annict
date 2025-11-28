# typed: false
# frozen_string_literal: true

RSpec.describe "PATCH /settings/email_notification", type: :request do
  it "ログイン済みユーザーは通知設定を更新できること" do
    user = create(:registered_user)
    login_as(user, scope: :user)

    patch "/settings/email_notification", params: {
      email_notification: {
        event_followed_user: "0",
        event_liked_episode_record: "1",
        event_favorite_works_added: "0",
        event_related_works_added: "1"
      }
    }

    expect(response).to redirect_to(settings_email_notification_path)
    expect(flash[:notice]).to eq("更新しました")

    # 設定が更新されていることを確認
    email_notification = user.reload.email_notification
    expect(email_notification.event_followed_user).to be false
    expect(email_notification.event_liked_episode_record).to be true
    expect(email_notification.event_favorite_works_added).to be false
    expect(email_notification.event_related_works_added).to be true
  end

  it "未ログインユーザーはログインページにリダイレクトされること" do
    patch "/settings/email_notification", params: {
      email_notification: {
        event_followed_user: "1"
      }
    }

    expect(response).to redirect_to(new_user_session_path)
  end

  it "全ての通知をオフに設定できること" do
    user = create(:registered_user)
    login_as(user, scope: :user)

    # 最初は全てオンになっていることを確認
    email_notification = user.email_notification
    expect(email_notification.event_followed_user).to be true
    expect(email_notification.event_liked_episode_record).to be true
    expect(email_notification.event_favorite_works_added).to be true
    expect(email_notification.event_related_works_added).to be true

    patch "/settings/email_notification", params: {
      email_notification: {
        event_followed_user: "0",
        event_liked_episode_record: "0",
        event_favorite_works_added: "0",
        event_related_works_added: "0"
      }
    }

    expect(response).to redirect_to(settings_email_notification_path)

    # 全てオフになっていることを確認
    email_notification.reload
    expect(email_notification.event_followed_user).to be false
    expect(email_notification.event_liked_episode_record).to be false
    expect(email_notification.event_favorite_works_added).to be false
    expect(email_notification.event_related_works_added).to be false
  end

  it "一部の通知設定のみ更新できること" do
    user = create(:registered_user)
    login_as(user, scope: :user)

    patch "/settings/email_notification", params: {
      email_notification: {
        event_followed_user: "0",
        event_favorite_works_added: "1"
      }
    }

    expect(response).to redirect_to(settings_email_notification_path)

    email_notification = user.reload.email_notification
    expect(email_notification.event_followed_user).to be false
    expect(email_notification.event_liked_episode_record).to be true # デフォルト値のまま
    expect(email_notification.event_favorite_works_added).to be true
    expect(email_notification.event_related_works_added).to be true # デフォルト値のまま
  end

  it "許可されていないパラメータは無視されること" do
    user = create(:registered_user)
    login_as(user, scope: :user)

    patch "/settings/email_notification", params: {
      email_notification: {
        event_followed_user: "0",
        event_liked_episode_record: "1",
        event_favorite_works_added: "0",
        event_related_works_added: "1",
        malicious_param: "evil",
        user_id: 9999
      }
    }

    expect(response).to redirect_to(settings_email_notification_path)

    # 許可されたパラメータのみが更新される
    email_notification = user.reload.email_notification
    expect(email_notification.event_followed_user).to be false
    expect(email_notification.event_liked_episode_record).to be true
    expect(email_notification.event_favorite_works_added).to be false
    expect(email_notification.event_related_works_added).to be true
    expect(email_notification.user_id).to eq(user.id) # user_idは変更されない
  end

  it "空のパラメータでも正しく処理されること" do
    user = create(:registered_user)
    login_as(user, scope: :user)

    # 空のパラメータを送信（全てのチェックボックスがオフになる想定）
    patch "/settings/email_notification", params: {
      email_notification: {
        # Railsのチェックボックスヘルパーは隠しフィールドで "0" を送信する
        event_followed_user: "0",
        event_liked_episode_record: "0",
        event_favorite_works_added: "0",
        event_related_works_added: "0"
      }
    }

    expect(response).to redirect_to(settings_email_notification_path)

    # 全てオフになっていることを確認
    email_notification = user.reload.email_notification
    expect(email_notification.event_followed_user).to be false
    expect(email_notification.event_liked_episode_record).to be false
    expect(email_notification.event_favorite_works_added).to be false
    expect(email_notification.event_related_works_added).to be false
  end

  it "更新に失敗した場合はshowビューが表示されること" do
    user = create(:registered_user)
    login_as(user, scope: :user)

    # EmailNotificationのupdateメソッドをスタブして失敗させる
    email_notification = user.email_notification
    allow(EmailNotification).to receive(:find_by).and_return(email_notification)
    allow(email_notification).to receive(:update).and_return(false)

    patch "/settings/email_notification", params: {
      email_notification: {
        event_followed_user: "0"
      }
    }

    expect(response.status).to eq(200)
    expect(response.body).to include("メール通知")
    expect(flash[:notice]).to be_nil
  end
end
