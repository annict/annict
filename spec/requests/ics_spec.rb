# typed: strict
# frozen_string_literal: true

RSpec.describe "GET /@:username/ics", type: :request do
  it "ユーザーが存在する場合、ICカレンダー形式で応答すること" do
    user = FactoryBot.create(:user, username: "test_user")
    work = FactoryBot.create(:work, started_on: Date.today)
    program = FactoryBot.create(:program, work:)
    episode = FactoryBot.create(:episode, work:)
    slot = FactoryBot.create(:slot, program:, episode:, started_at: 1.hour.since)
    
    status = FactoryBot.create(:status, user:, work:, kind: :wanna_watch)
    FactoryBot.create(:library_entry, user:, program:, work:, status:)

    get "/@#{user.username}/ics"

    expect(response).to have_http_status(:ok)
    expect(response.content_type).to eq("text/calendar")
    expect(response.headers["Content-Disposition"]).to include('filename="annict.ics"')
    expect(response.body).to include("BEGIN:VCALENDAR")
    expect(response.body).to include("END:VCALENDAR")
  end

  it "ユーザーが存在しない場合、404エラーを返すこと" do
    expect {
      get "/@nonexistent_user/ics"
    }.to raise_error(ActiveRecord::RecordNotFound)
  end

  it "削除されたユーザーにアクセスした場合、404エラーを返すこと" do
    user = FactoryBot.create(:user, username: "deleted_user")
    user.update!(deleted_at: Time.zone.now)

    expect {
      get "/@#{user.username}/ics"
    }.to raise_error(ActiveRecord::RecordNotFound)
  end

  it "視聴リストに追加済みのアニメがない場合、空のカレンダーを返すこと" do
    user = FactoryBot.create(:user, username: "empty_user")

    get "/@#{user.username}/ics"

    expect(response).to have_http_status(:ok)
    expect(response.content_type).to eq("text/calendar")
    expect(response.body).to include("BEGIN:VCALENDAR")
    expect(response.body).to include("END:VCALENDAR")
  end

  it "過去の放送枠は含まれないこと" do
    user = FactoryBot.create(:user, username: "test_user")
    work = FactoryBot.create(:work)
    program = FactoryBot.create(:program, work:)
    episode = FactoryBot.create(:episode, work:)
    slot = FactoryBot.create(:slot, program:, episode:, started_at: 1.day.ago)
    
    status = FactoryBot.create(:status, user:, work:, kind: :watching)
    FactoryBot.create(:library_entry, user:, program:, work:, status:)

    get "/@#{user.username}/ics"

    expect(response).to have_http_status(:ok)
    expect(response.body).to include("BEGIN:VCALENDAR")
    expect(response.body).to include("END:VCALENDAR")
  end

  it "8日以降の放送枠は含まれないこと" do
    user = FactoryBot.create(:user, username: "test_user")
    work = FactoryBot.create(:work)
    program = FactoryBot.create(:program, work:)
    episode = FactoryBot.create(:episode, work:)
    slot = FactoryBot.create(:slot, program:, episode:, started_at: 8.days.since)
    
    status = FactoryBot.create(:status, user:, work:, kind: :watching)
    FactoryBot.create(:library_entry, user:, program:, work:, status:)

    get "/@#{user.username}/ics"

    expect(response).to have_http_status(:ok)
    expect(response.body).to include("BEGIN:VCALENDAR")
    expect(response.body).to include("END:VCALENDAR")
  end
end