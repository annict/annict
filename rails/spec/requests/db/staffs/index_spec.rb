# typed: false
# frozen_string_literal: true

RSpec.describe "GET /db/works/:work_id/staffs", type: :request do
  it "ユーザーがログインしていない場合、スタッフ一覧が表示されること" do
    staff = FactoryBot.create(:staff)

    get "/db/works/#{staff.work_id}/staffs"

    expect(response.status).to eq(200)
    expect(response.body).to include(staff.resource.name)
  end

  it "ユーザーがログインしている場合、スタッフ一覧が表示されること" do
    user = FactoryBot.create(:registered_user)
    staff = FactoryBot.create(:staff)
    login_as(user, scope: :user)

    get "/db/works/#{staff.work_id}/staffs"

    expect(response.status).to eq(200)
    expect(response.body).to include(staff.resource.name)
  end

  it "削除されたWorkの場合、404エラーが返されること" do
    work = FactoryBot.create(:work, deleted_at: Time.current)
    FactoryBot.create(:staff, work: work)

    expect {
      get "/db/works/#{work.id}/staffs"
    }.to raise_error(ActiveRecord::RecordNotFound)
  end

  it "存在しないWorkの場合、404エラーが返されること" do
    expect {
      get "/db/works/99999/staffs"
    }.to raise_error(ActiveRecord::RecordNotFound)
  end

  it "複数のスタッフが存在する場合、sort_number順に表示されること" do
    work = FactoryBot.create(:work)
    staff1 = FactoryBot.create(:staff, work: work, sort_number: 20)
    staff2 = FactoryBot.create(:staff, work: work, sort_number: 10)
    staff3 = FactoryBot.create(:staff, work: work, sort_number: 30)

    get "/db/works/#{work.id}/staffs"

    expect(response.status).to eq(200)
    # sort_number順に並んでいることを確認
    expect(response.body.index(staff2.resource.name)).to be < response.body.index(staff1.resource.name)
    expect(response.body.index(staff1.resource.name)).to be < response.body.index(staff3.resource.name)
  end

  it "削除されたスタッフは表示されないこと" do
    work = FactoryBot.create(:work)
    staff1 = FactoryBot.create(:staff, work: work)
    staff2 = FactoryBot.create(:staff, work: work, deleted_at: Time.current)

    get "/db/works/#{work.id}/staffs"

    expect(response.status).to eq(200)
    expect(response.body).to include(staff1.resource.name)
    expect(response.body).not_to include(staff2.resource.name)
  end
end
