# typed: false
# frozen_string_literal: true

RSpec.describe "GET /r/:provider/:url_hash", type: :request do
  it "TwitterのURL hashが存在するとき、エピソード記録のページにリダイレクトすること" do
    user = FactoryBot.create(:user, :with_profile)
    work = FactoryBot.create(:work)
    episode = FactoryBot.create(:episode, work:)
    record = FactoryBot.create(:record, user:, work:)
    episode_record = FactoryBot.create(:episode_record, user:, work:, episode:, record:, twitter_url_hash: "abcdefghij")

    get "/r/tw/abcdefghij"

    expect(response).to have_http_status(:moved_permanently)
    expect(response).to redirect_to(episode_record.share_url_with_query(:twitter))
  end

  it "FacebookのURL hashが存在するとき、エピソード記録のページにリダイレクトすること" do
    user = FactoryBot.create(:user, :with_profile)
    work = FactoryBot.create(:work)
    episode = FactoryBot.create(:episode, work:)
    record = FactoryBot.create(:record, user:, work:)
    episode_record = FactoryBot.create(:episode_record, user:, work:, episode:, record:, facebook_url_hash: "abcdefghij")

    get "/r/fb/abcdefghij"

    expect(response).to have_http_status(:moved_permanently)
    expect(response).to redirect_to(episode_record.share_url_with_query(:facebook))
  end

  it "TwitterのURL hashが存在しないとき、404エラーが返されることエラーが発生すること" do
    get "/r/tw/not_found1"

    expect(response.status).to eq(404)
  end

  it "FacebookのURL hashが存在しないとき、404エラーが返されることエラーが発生すること" do
    get "/r/fb/not_found1"

    expect(response.status).to eq(404)
  end

  it "削除されたエピソード記録のとき、404エラーが返されることエラーが発生すること" do
    user = FactoryBot.create(:user, :with_profile)
    work = FactoryBot.create(:work)
    episode = FactoryBot.create(:episode, work:)
    record = FactoryBot.create(:record, user:, work:)
    episode_record = FactoryBot.create(:episode_record, user:, work:, episode:, record:, twitter_url_hash: "deletedabc")
    episode_record.update!(deleted_at: Time.current)

    get "/r/tw/deletedabc"

    expect(response.status).to eq(404)
  end

  it "\u4e0d\u660e\u306eprovider\u304c\u6307\u5b9a\u3055\u308c\u305f\u3068\u304d\u3001\u30eb\u30fc\u30c6\u30a3\u30f3\u30b0\u30a8\u30e9\u30fc\u304c\u767a\u751f\u3059\u308b\u3053\u3068" do
    get "/r/unknown/abcdefghij"

    expect(response.status).to eq(404)
  end

  it "\u7121\u52b9\u306aURL hash\u5f62\u5f0f\u306e\u3068\u304d\u3001404\u30a8\u30e9\u30fc\u304c\u767a\u751f\u3059\u308b\u3053\u3068" do
    get "/r/tw/invalid"

    expect(response.status).to eq(404)
  end
end
