# typed: false
# frozen_string_literal: true

RSpec.describe "PATCH /settings/options", type: :request do
  it "ログイン済みユーザーがオプションを更新できること" do
    user = create(:registered_user)
    login_as(user, scope: :user)

    expect(user.hide_record_body?).to eq(true)

    patch "/settings/options", params: {setting: {hide_record_body: false}}

    expect(response.status).to eq(302)
    expect(flash[:notice]).to eq("更新しました")

    user.reload
    expect(user.hide_record_body?).to eq(false)
  end

  it "未ログインユーザーはアクセスできないこと" do
    patch "/settings/options", params: {setting: {hide_record_body: false}}

    expect(response.status).to eq(302)
    expect(response).to redirect_to(new_user_session_path)
  end
end
