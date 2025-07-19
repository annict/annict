# typed: false
# frozen_string_literal: true

RSpec.describe "POST /api/internal/library_entries", type: :request do
  it "未ログイン時は空の配列を返すこと" do
    post "/api/internal/library_entries"

    expect(response.status).to eq(200)
    expect(JSON.parse(response.body)).to eq([])
  end

  it "work_idsパラメータがない場合は空の配列を返すこと" do
    user = FactoryBot.create(:user, :with_email_notification)

    login_as(user, scope: :user)
    post "/api/internal/library_entries"

    expect(response.status).to eq(200)
    expect(JSON.parse(response.body)).to eq([])
  end

  it "該当するlibrary_entriesがない場合は空のハッシュを返すこと" do
    user = FactoryBot.create(:user, :with_email_notification)
    work = FactoryBot.create(:work)

    login_as(user, scope: :user)
    post "/api/internal/library_entries", params: {work_ids: work.id.to_s}

    expect(response.status).to eq(200)
    expect(JSON.parse(response.body)).to eq({})
  end

  it "library_entriesがある場合はstatus_kindsを返すこと" do
    user = FactoryBot.create(:user, :with_email_notification)
    work1 = FactoryBot.create(:work)
    work2 = FactoryBot.create(:work)
    status1 = FactoryBot.create(:status, user: user, work: work1, kind: :watching)
    status2 = FactoryBot.create(:status, user: user, work: work2, kind: :watched)
    FactoryBot.create(:library_entry, user: user, work: work1, status: status1)
    FactoryBot.create(:library_entry, user: user, work: work2, status: status2)

    login_as(user, scope: :user)
    post "/api/internal/library_entries", params: {work_ids: "#{work1.id},#{work2.id}"}

    expect(response.status).to eq(200)
    response_body = JSON.parse(response.body)
    expect(response_body[work1.id.to_s]).to eq("watching")
    expect(response_body[work2.id.to_s]).to eq("completed")
  end

  it "複数のwork_idsに対して正しいstatus_kindsを返すこと" do
    user = FactoryBot.create(:user, :with_email_notification)
    work1 = FactoryBot.create(:work)
    work2 = FactoryBot.create(:work)
    work3 = FactoryBot.create(:work)
    status1 = FactoryBot.create(:status, user: user, work: work1, kind: :wanna_watch)
    status2 = FactoryBot.create(:status, user: user, work: work2, kind: :on_hold)
    FactoryBot.create(:library_entry, user: user, work: work1, status: status1)
    FactoryBot.create(:library_entry, user: user, work: work2, status: status2)

    login_as(user, scope: :user)
    post "/api/internal/library_entries", params: {work_ids: "#{work1.id},#{work2.id},#{work3.id}"}

    expect(response.status).to eq(200)
    response_body = JSON.parse(response.body)
    expect(response_body[work1.id.to_s]).to eq("plan_to_watch")
    expect(response_body[work2.id.to_s]).to eq("on_hold")
    expect(response_body).not_to have_key(work3.id.to_s)
  end

  it "他のユーザーのlibrary_entriesは結果に含まれないこと" do
    user = FactoryBot.create(:user, :with_email_notification)
    other_user = FactoryBot.create(:user, :with_email_notification)
    work = FactoryBot.create(:work)
    status = FactoryBot.create(:status, user: other_user, work: work, kind: :watching)
    FactoryBot.create(:library_entry, user: other_user, work: work, status: status)

    login_as(user, scope: :user)
    post "/api/internal/library_entries", params: {work_ids: work.id.to_s}

    expect(response.status).to eq(200)
    expect(JSON.parse(response.body)).to eq({})
  end
end
