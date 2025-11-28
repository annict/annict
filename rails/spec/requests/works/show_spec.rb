# typed: false
# frozen_string_literal: true

RSpec.describe "GET /works/:work_id", type: :request do
  it "ユーザーがログインしていないとき、作品情報が表示されること" do
    work = create(:work)

    get "/works/#{work.id}"

    expect(response.status).to eq(200)
    expect(response.body).to include(work.title)
  end

  it "ユーザーがログインしているとき、作品情報が表示されること" do
    user = create(:registered_user)
    work = create(:work)
    login_as(user, scope: :user)

    get "/works/#{work.id}"

    expect(response.status).to eq(200)
    expect(response.body).to include(work.title)
  end

  it "存在しない作品にアクセスしたとき、RecordNotFoundエラーが発生すること" do
    expect {
      get "/works/999999"
    }.to raise_error(ActiveRecord::RecordNotFound)
  end

  it "削除された作品にアクセスしたとき、RecordNotFoundエラーが発生すること" do
    work = create(:work, deleted_at: Time.current)

    expect {
      get "/works/#{work.id}"
    }.to raise_error(ActiveRecord::RecordNotFound)
  end

  it "トレーラーが追加されているとき、トレーラーのタイトルが表示されること" do
    work = create(:work)
    trailer = create(:trailer, work: work)

    get "/works/#{work.id}"

    expect(response.status).to eq(200)
    expect(response.body).to include(trailer.title)
  end

  it "エピソードが追加されているとき、エピソードのタイトルが表示されること" do
    work = create(:work)
    episode = create(:episode, work: work)

    get "/works/#{work.id}"

    expect(response.status).to eq(200)
    expect(response.body).to include(episode.title)
  end

  it "VODが追加されているとき、VODサービスへのリンクが表示されること" do
    work = create(:work)
    channel = Channel.with_vod.first
    program = create(:program, work: work, channel: channel, vod_title_code: "xxx")
    vod_title_url = "https://example.com/#{program.vod_title_code}"
    allow_any_instance_of(Program).to receive(:vod_title_url).and_return(vod_title_url)

    get "/works/#{work.id}"

    expect(response.status).to eq(200)
    expect(response.body).to include(vod_title_url)
  end

  it "作品記録が追加されているとき、作品記録の本文が表示されること" do
    work = create(:work)
    record = create(:record, work: work)
    work_record = create(:work_record, work: work, record: record)

    get "/works/#{work.id}"

    expect(response.status).to eq(200)
    expect(response.body).to include(work_record.body)
  end
end
