# typed: false
# frozen_string_literal: true

RSpec.describe "GET /fragment/@:username/records/:record_id", type: :request do
  it "ユーザーが存在しない場合、404エラーを返すこと" do
    get "/fragment/@nonexistentuser/records/123"
    
    expect(response.status).to eq(404)
  end

  it "レコードが存在しない場合、404エラーを返すこと" do
    user = FactoryBot.create(:registered_user)

    get "/fragment/@#{user.username}/records/nonexistent"
    
    expect(response.status).to eq(404)
  end

  it "削除されたユーザーの記録を表示しようとした場合、404エラーを返すこと" do
    user = FactoryBot.create(:registered_user)
    work = FactoryBot.create(:work)
    record = FactoryBot.create(:record, user: user, work: work)
    user.destroy!

    get "/fragment/@#{user.username}/records/#{record.id}"
    
    expect(response.status).to eq(404)
  end

  it "削除された記録を表示しようとした場合、404エラーを返すこと" do
    user = FactoryBot.create(:registered_user)
    work = FactoryBot.create(:work)
    record = FactoryBot.create(:record, user: user, work: work)
    record.destroy!

    get "/fragment/@#{user.username}/records/#{record.id}"
    
    expect(response.status).to eq(404)
  end

  it "別のユーザーの記録にアクセスした場合でも、記録が表示されること" do
    owner = FactoryBot.create(:registered_user)
    viewer = FactoryBot.create(:registered_user)
    work = FactoryBot.create(:work)
    record = FactoryBot.create(:record, user: owner, work: work)

    login_as(viewer, scope: :user)

    get "/fragment/@#{owner.username}/records/#{record.id}"

    expect(response).to have_http_status(:ok)
  end

  it "未ログインでも記録が表示されること" do
    user = FactoryBot.create(:registered_user)
    work = FactoryBot.create(:work)
    record = FactoryBot.create(:record, user: user, work: work)

    get "/fragment/@#{user.username}/records/#{record.id}"

    expect(response).to have_http_status(:ok)
  end
end
