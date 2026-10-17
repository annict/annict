# typed: false
# frozen_string_literal: true

RSpec.describe "GET /@:username/favorite_people", type: :request do
  it "お気に入りがいないとき、アクセスできること" do
    user = create(:registered_user)
    get "/@#{user.username}/favorite_people"

    expect(response.status).to eq(200)
    expect(response.body).to include("人物はいません")
  end

  it "お気に入り人物がいるとき、正常にアクセスできること" do
    user = create(:registered_user)
    person = create(:person)
    create(:person_favorite, user: user, person: person)

    get "/@#{user.username}/favorite_people"

    expect(response.status).to eq(200)
  end

  it "人物が非公開・削除済みのお気に入りは表示されないこと" do
    user = create(:registered_user)
    published_person = create(:person, name: "公開中の声優")
    unpublished_person = create(:person, :unpublished, name: "非公開の声優")
    deleted_person = create(:person, :deleted, name: "削除済みのスタッフ")
    create(:cast, person: published_person, name: published_person.name)
    create(:cast, person: unpublished_person, name: unpublished_person.name)
    create(:staff, resource: deleted_person, name: deleted_person.name)
    create(:person_favorite, user:, person: published_person)
    create(:person_favorite, user:, person: unpublished_person)
    create(:person_favorite, user:, person: deleted_person)

    get "/@#{user.username}/favorite_people"

    expect(response.status).to eq(200)
    expect(response.body).to include("公開中の声優")
    expect(response.body).not_to include("非公開の声優")
    expect(response.body).not_to include("削除済みのスタッフ")
  end
end
