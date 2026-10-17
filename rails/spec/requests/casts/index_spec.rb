# typed: false
# frozen_string_literal: true

RSpec.describe "GET /works/:work_id/casts", type: :request do
  it "キャスト一覧が表示されること" do
    work = FactoryBot.create(:work)
    person = FactoryBot.create(:person, name: "公開中の声優")
    FactoryBot.create(:cast, work:, person:)

    get "/works/#{work.id}/casts"

    expect(response.status).to eq(200)
    expect(response.body).to include("公開中の声優")
  end

  it "人物が非公開・削除済みのキャストは表示されないこと" do
    work = FactoryBot.create(:work)
    published_person = FactoryBot.create(:person, name: "公開中の声優")
    unpublished_person = FactoryBot.create(:person, :unpublished, name: "非公開の声優")
    deleted_person = FactoryBot.create(:person, :deleted, name: "削除済みの声優")
    FactoryBot.create(:cast, work:, person: published_person, name: published_person.name)
    FactoryBot.create(:cast, work:, person: unpublished_person, name: unpublished_person.name)
    FactoryBot.create(:cast, work:, person: deleted_person, name: deleted_person.name)

    get "/works/#{work.id}/casts"

    expect(response.status).to eq(200)
    expect(response.body).to include("公開中の声優")
    expect(response.body).not_to include("非公開の声優")
    expect(response.body).not_to include("削除済みの声優")
  end
end
