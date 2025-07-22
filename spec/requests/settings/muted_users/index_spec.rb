# typed: false
# frozen_string_literal: true

RSpec.describe "GET /settings/muted_users", type: :request do
  it "未ログイン時、ログインページにリダイレクトすること" do
    get "/settings/muted_users"

    expect(response).to redirect_to(new_user_session_path)
    expect(flash[:alert]).to eq("ログインしてください")
  end

  it "ミュートしたユーザーがいない場合、空の状態が表示されること" do
    user = create(:registered_user)
    login_as(user, scope: :user)

    get "/settings/muted_users"

    expect(response.status).to eq(200)
    expect(response.body).to include("ミュートした人はいません")
  end

  it "ミュートしたユーザーが1人いる場合、そのユーザーが表示されること" do
    user_1 = create(:registered_user)
    user_2 = create(:registered_user)
    user_1.mute(user_2)
    login_as(user_1, scope: :user)

    get "/settings/muted_users"

    expect(response.status).to eq(200)
    expect(response.body).to include(user_2.profile.name)
    expect(response.body).to include("ミュートを解除")
  end

  it "複数のミュートしたユーザーがいる場合、新しい順に表示されること" do
    user_1 = create(:registered_user)
    user_2 = create(:registered_user)
    user_3 = create(:registered_user)

    user_1.mute(user_2)
    user_1.mute(user_3)
    login_as(user_1, scope: :user)

    get "/settings/muted_users"

    expect(response.status).to eq(200)
    expect(response.body).to include(user_2.profile.name)
    expect(response.body).to include(user_3.profile.name)

    # user_3が後にミュートされたので先に表示される（新しい順）
    user_2_position = response.body.index(user_2.profile.name)
    user_3_position = response.body.index(user_3.profile.name)
    expect(user_3_position).to be < user_2_position
  end
end
