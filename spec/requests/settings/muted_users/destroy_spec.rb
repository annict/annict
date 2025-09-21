# typed: false
# frozen_string_literal: true

RSpec.describe "DELETE /settings/muted_users/:mute_user_id", type: :request do
  it "ログインしているとき、ミュートを解除してリダイレクトすること" do
    user_1 = FactoryBot.create(:registered_user)
    user_2 = FactoryBot.create(:registered_user)
    user_1.mute(user_2)
    mute_user = user_1.mute_users.find_by(muted_user: user_2)

    login_as(user_1, scope: :user)

    delete "/settings/muted_users/#{mute_user.id}"

    expect(response.status).to eq(302)
    expect(response).to redirect_to(settings_muted_user_list_path)
    expect(flash[:notice]).to eq("ミュートを解除しました")
    expect(user_1.mute_users.exists?(id: mute_user.id)).to be false
  end

  it "ログインしていないとき、ログインページにリダイレクトすること" do
    user_1 = FactoryBot.create(:registered_user)
    user_2 = FactoryBot.create(:registered_user)
    user_1.mute(user_2)
    mute_user = user_1.mute_users.find_by(muted_user: user_2)

    delete "/settings/muted_users/#{mute_user.id}"

    expect(response.status).to eq(302)
    expect(response).to redirect_to(new_user_session_path)
  end

  it "他のユーザーのミュート情報を削除しようとしたとき、404エラーが返されること" do
    user_1 = FactoryBot.create(:registered_user)
    user_2 = FactoryBot.create(:registered_user)
    user_3 = FactoryBot.create(:registered_user)
    user_1.mute(user_2)
    mute_user = user_1.mute_users.find_by(muted_user: user_2)

    login_as(user_3, scope: :user)

    delete "/settings/muted_users/#{mute_user.id}"

    expect(response).to have_http_status(:not_found)
  end

  it "存在しないミュート情報を削除しようとしたとき、404エラーが返されること" do
    user_1 = FactoryBot.create(:registered_user)

    login_as(user_1, scope: :user)

    delete "/settings/muted_users/999999"

    expect(response).to have_http_status(:not_found)
  end
end
