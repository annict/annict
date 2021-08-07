# frozen_string_literal: true

describe "GET /notifications", type: :request do
  let!(:user_1) { create(:registered_user) }
  let!(:user_2) { create(:registered_user) }
  let!(:anime) { create(:anime) }
  let!(:record) { create(:record, user: user_1, anime: anime) }
  let!(:anime_record) { create(:anime_record, user: user_1, anime: anime, record: record) }

  before do
    Creators::LikeCreator.new(user: user_2, likeable: record).call

    login_as(user_1, scope: :user)
  end

  it "通知が表示されること" do
    get "/notifications"

    expect(response.status).to eq(200)
    expect(response.body).to include(user_2.profile.name)
    expect(response.body).to include("に「いいね！」しました")
  end
end
