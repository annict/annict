# typed: false
# frozen_string_literal: true

RSpec.describe "GET /settings/email_notification", type: :request do
  it "ログイン済みユーザーはページが表示されること" do
    user = create(:registered_user)
    login_as(user, scope: :user)

    get "/settings/email_notification"

    expect(response.status).to eq(200)
    expect(response.body).to include("メール通知")
  end

  it "未ログインユーザーはログインページにリダイレクトされること" do
    get "/settings/email_notification"

    expect(response).to redirect_to(new_user_session_path)
  end

  it "ユーザーの現在の通知設定が表示されること" do
    user = create(:registered_user)
    email_notification = user.email_notification
    email_notification.update!(
      event_followed_user: true,
      event_liked_episode_record: false,
      event_favorite_works_added: true,
      event_related_works_added: false
    )
    login_as(user, scope: :user)

    get "/settings/email_notification"

    expect(response.status).to eq(200)

    # チェックボックスの存在と状態を確認
    # event_followed_user: true
    expect(response.body).to include('type="checkbox" value="1" checked="checked" name="email_notification[event_followed_user]"')
    # event_liked_episode_record: false
    expect(response.body).to include('type="checkbox" value="1" name="email_notification[event_liked_episode_record]"')
    expect(response.body).not_to include('checked="checked" name="email_notification[event_liked_episode_record]"')
    # event_favorite_works_added: true
    expect(response.body).to include('type="checkbox" value="1" checked="checked" name="email_notification[event_favorite_works_added]"')
    # event_related_works_added: false
    expect(response.body).to include('type="checkbox" value="1" name="email_notification[event_related_works_added]"')
    expect(response.body).not_to include('checked="checked" name="email_notification[event_related_works_added]"')
  end

  it "通知設定のフォームが表示されること" do
    user = create(:registered_user)
    login_as(user, scope: :user)

    get "/settings/email_notification"

    expect(response.status).to eq(200)
    # フォーム要素の存在を確認
    expect(response.body).to include("<form")
    expect(response.body).to match(%r{action="/settings/email_notification})
    expect(response.body).to include('method="post"')
    expect(response.body).to include('name="email_notification[event_followed_user]"')
    expect(response.body).to include('name="email_notification[event_liked_episode_record]"')
    expect(response.body).to include('name="email_notification[event_favorite_works_added]"')
    expect(response.body).to include('name="email_notification[event_related_works_added]"')
    expect(response.body).to include('type="submit"')
  end
end
