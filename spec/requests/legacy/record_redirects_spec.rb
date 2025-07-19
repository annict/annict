# typed: false
# frozen_string_literal: true

RSpec.describe "GET /checkins/redirect/:provider/:url_hash", type: :request do
  it "TwitterのURL hashが存在するとき、エピソード記録のページにリダイレクトすること" do
    user = FactoryBot.create(:user, :with_profile)
    work = FactoryBot.create(:work)
    episode = FactoryBot.create(:episode, work:)
    record = FactoryBot.create(:record, user:, work:)
    episode_record = FactoryBot.create(:episode_record, user:, work:, episode:, record:, twitter_url_hash: "abcdefghij")

    # share_url_with_queryメソッドをスタブ化
    allow(episode_record).to receive(:share_url_with_query).with(:twitter).and_return(episode_record.share_url)
    allow(EpisodeRecord).to receive_message_chain(:only_kept, :find_by!).and_return(episode_record)

    get "/checkins/redirect/tw/abcdefghij"

    expect(response).to have_http_status(:moved_permanently)
    expect(response).to redirect_to(episode_record.share_url)
  end

  it "FacebookのURL hashが存在するとき、エピソード記録のページにリダイレクトすること" do
    user = FactoryBot.create(:user, :with_profile)
    work = FactoryBot.create(:work)
    episode = FactoryBot.create(:episode, work:)
    record = FactoryBot.create(:record, user:, work:)
    episode_record = FactoryBot.create(:episode_record, user:, work:, episode:, record:, facebook_url_hash: "abcdefghij")

    # share_url_with_queryメソッドをスタブ化
    allow(episode_record).to receive(:share_url_with_query).with(:facebook).and_return(episode_record.share_url)
    allow(EpisodeRecord).to receive_message_chain(:only_kept, :find_by!).and_return(episode_record)

    get "/checkins/redirect/fb/abcdefghij"

    expect(response).to have_http_status(:moved_permanently)
    expect(response).to redirect_to(episode_record.share_url)
  end

  it "TwitterのURL hashが存在しないとき、ActiveRecord::RecordNotFoundエラーが発生すること" do
    # only_keptスコープでfind_by!が呼ばれるときにRecordNotFoundをレイズ
    allow(EpisodeRecord).to receive_message_chain(:only_kept, :find_by!).and_raise(ActiveRecord::RecordNotFound)

    expect do
      get "/checkins/redirect/tw/not_found1"
    end.to raise_error(ActiveRecord::RecordNotFound)
  end

  it "FacebookのURL hashが存在しないとき、ActiveRecord::RecordNotFoundエラーが発生すること" do
    # only_keptスコープでfind_by!が呼ばれるときにRecordNotFoundをレイズ
    allow(EpisodeRecord).to receive_message_chain(:only_kept, :find_by!).and_raise(ActiveRecord::RecordNotFound)

    expect do
      get "/checkins/redirect/fb/not_found1"
    end.to raise_error(ActiveRecord::RecordNotFound)
  end

  it "削除されたエピソード記録のとき、ActiveRecord::RecordNotFoundエラーが発生すること" do
    user = FactoryBot.create(:user, :with_profile)
    work = FactoryBot.create(:work)
    episode = FactoryBot.create(:episode, work:)
    record = FactoryBot.create(:record, user:, work:)
    episode_record = FactoryBot.create(:episode_record, user:, work:, episode:, record:, twitter_url_hash: "deletedabc")
    episode_record.update!(deleted_at: Time.current)

    expect do
      get "/checkins/redirect/tw/deletedabc"
    end.to raise_error(ActiveRecord::RecordNotFound)
  end
end
