# typed: false
# frozen_string_literal: true

RSpec.describe "DELETE /v1/me/reviews/:id", type: :request do
  it "正常にレビューを削除できること" do
    user = create(:user, :with_profile, :with_setting)
    application = create(:oauth_application, owner: user)
    access_token = create(:oauth_access_token, application: application)
    work = create(:work, :with_current_season)
    create(:record, work: work, user: user)
    work_record = create(:work_record, work: work, user: user)

    expect(access_token.owner.work_records.count).to eq(1)

    delete api("/v1/me/reviews/#{work_record.id}", access_token: access_token.token)

    expect(response.status).to eq(204)
    expect(access_token.owner.work_records.count).to eq(0)
  end

  it "認証トークンがない場合、エラーが返されること" do
    user = create(:user, :with_profile, :with_setting)
    work = create(:work, :with_current_season)
    create(:record, work: work, user: user)
    work_record = create(:work_record, work: work, user: user)

    delete api("/v1/me/reviews/#{work_record.id}")

    expect(response.status).to eq(401)
  end

  it "存在しないレビューIDの場合、エラーが返されること" do
    user = create(:user, :with_profile, :with_setting)
    application = create(:oauth_application, owner: user)
    access_token = create(:oauth_access_token, application: application)

    delete api("/v1/me/reviews/invalid_id", access_token: access_token.token)

    expect(response.status).to eq(400)
  end

  it "他のユーザーのレビューを削除しようとした場合、エラーが返されること" do
    user1 = create(:user, :with_profile, :with_setting)
    user2 = create(:user, :with_profile, :with_setting)
    application = create(:oauth_application, owner: user2)
    access_token = create(:oauth_access_token, application: application)
    work = create(:work, :with_current_season)
    create(:record, work: work, user: user1)
    work_record = create(:work_record, work: work, user: user1)

    delete api("/v1/me/reviews/#{work_record.id}", access_token: access_token.token)

    expect(response.status).to eq(404)
  end
end
