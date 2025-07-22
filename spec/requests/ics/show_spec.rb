# typed: false
# frozen_string_literal: true

RSpec.describe "GET /ics", type: :request do
  it "ユーザーが存在する場合、iCalendar形式で応答すること" do
    user = FactoryBot.create(:user, username: "test_user")
    work = FactoryBot.create(:work, started_on: Time.zone.today)
    program = FactoryBot.create(:program, work:)
    episode = FactoryBot.create(:episode, work:)
    FactoryBot.create(:slot, program:, episode:, started_at: 1.hour.since)

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
    FactoryBot.create(:slot, program:, episode:, started_at: 1.day.ago)

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
    FactoryBot.create(:slot, program:, episode:, started_at: 8.days.since)

    status = FactoryBot.create(:status, user:, work:, kind: :watching)
    FactoryBot.create(:library_entry, user:, program:, work:, status:)

    get "/@#{user.username}/ics"

    expect(response).to have_http_status(:ok)
    expect(response.body).to include("BEGIN:VCALENDAR")
    expect(response.body).to include("END:VCALENDAR")
  end

  it "視聴済みのエピソードは含まれないこと" do
    user = FactoryBot.create(:user, username: "test_user")
    work = FactoryBot.create(:work)
    program = FactoryBot.create(:program, work:)
    episode1 = FactoryBot.create(:episode, work:)
    episode2 = FactoryBot.create(:episode, work:)
    FactoryBot.create(:slot, program:, episode: episode1, started_at: 1.hour.since)
    FactoryBot.create(:slot, program:, episode: episode2, started_at: 2.hours.since)

    status = FactoryBot.create(:status, user:, work:, kind: :watching)
    FactoryBot.create(:library_entry, user:, program:, work:, status:, watched_episode_ids: [episode1.id])

    get "/@#{user.username}/ics"

    expect(response).to have_http_status(:ok)
    expect(response.body).to include("BEGIN:VCALENDAR")
    expect(response.body).to include("END:VCALENDAR")
  end

  it "視聴予定（wanna_watch）のアニメの放送枠が含まれること" do
    user = FactoryBot.create(:user, username: "test_user")
    work = FactoryBot.create(:work)
    program = FactoryBot.create(:program, work:)
    episode = FactoryBot.create(:episode, work:)
    FactoryBot.create(:slot, program:, episode:, started_at: 1.hour.since)

    status = FactoryBot.create(:status, user:, work:, kind: :wanna_watch)
    FactoryBot.create(:library_entry, user:, program:, work:, status:)

    get "/@#{user.username}/ics"

    expect(response).to have_http_status(:ok)
    expect(response.body).to include("BEGIN:VCALENDAR")
    expect(response.body).to include("END:VCALENDAR")
  end

  it "視聴中（watching）のアニメの放送枠が含まれること" do
    user = FactoryBot.create(:user, username: "test_user")
    work = FactoryBot.create(:work)
    program = FactoryBot.create(:program, work:)
    episode = FactoryBot.create(:episode, work:)
    FactoryBot.create(:slot, program:, episode:, started_at: 1.hour.since)

    status = FactoryBot.create(:status, user:, work:, kind: :watching)
    FactoryBot.create(:library_entry, user:, program:, work:, status:)

    get "/@#{user.username}/ics"

    expect(response).to have_http_status(:ok)
    expect(response.body).to include("BEGIN:VCALENDAR")
    expect(response.body).to include("END:VCALENDAR")
  end

  it "番組（program）が設定されていないライブラリエントリは無視されること" do
    user = FactoryBot.create(:user, username: "test_user")
    work = FactoryBot.create(:work)

    status = FactoryBot.create(:status, user:, work:, kind: :watching)
    FactoryBot.create(:library_entry, user:, program: nil, work:, status:)

    get "/@#{user.username}/ics"

    expect(response).to have_http_status(:ok)
    expect(response.body).to include("BEGIN:VCALENDAR")
    expect(response.body).to include("END:VCALENDAR")
  end

  it "開始日（started_on）が設定されているアニメが含まれること" do
    user = FactoryBot.create(:user, username: "test_user")
    work_with_start_date = FactoryBot.create(:work, started_on: Time.zone.today)
    work_without_start_date = FactoryBot.create(:work, started_on: nil)

    FactoryBot.create(:status, user:, work: work_with_start_date, kind: :watching)
    FactoryBot.create(:status, user:, work: work_without_start_date, kind: :watching)

    get "/@#{user.username}/ics"

    expect(response).to have_http_status(:ok)
    expect(response.body).to include("BEGIN:VCALENDAR")
    expect(response.body).to include("END:VCALENDAR")
  end

  it "削除済みの放送枠は含まれないこと" do
    user = FactoryBot.create(:user, username: "test_user")
    work = FactoryBot.create(:work)
    program = FactoryBot.create(:program, work:)
    episode = FactoryBot.create(:episode, work:)
    slot = FactoryBot.create(:slot, program:, episode:, started_at: 1.hour.since)
    slot.update!(deleted_at: Time.zone.now)

    status = FactoryBot.create(:status, user:, work:, kind: :watching)
    FactoryBot.create(:library_entry, user:, program:, work:, status:)

    get "/@#{user.username}/ics"

    expect(response).to have_http_status(:ok)
    expect(response.body).to include("BEGIN:VCALENDAR")
    expect(response.body).to include("END:VCALENDAR")
  end

  it "今日の0時以降7日後の23時59分59秒までの放送枠が含まれること" do
    user = FactoryBot.create(:user, username: "test_user")
    work = FactoryBot.create(:work)
    program = FactoryBot.create(:program, work:)
    episode_today = FactoryBot.create(:episode, work:)
    episode_7days = FactoryBot.create(:episode, work:)
    episode_8days = FactoryBot.create(:episode, work:)

    FactoryBot.create(:slot, program:, episode: episode_today, started_at: Time.zone.today.beginning_of_day + 1.hour)
    FactoryBot.create(:slot, program:, episode: episode_7days, started_at: 7.days.since.end_of_day - 1.hour)
    FactoryBot.create(:slot, program:, episode: episode_8days, started_at: 8.days.since.beginning_of_day)

    status = FactoryBot.create(:status, user:, work:, kind: :watching)
    FactoryBot.create(:library_entry, user:, program:, work:, status:)

    get "/@#{user.username}/ics"

    expect(response).to have_http_status(:ok)
    expect(response.body).to include("BEGIN:VCALENDAR")
    expect(response.body).to include("END:VCALENDAR")
  end
end
