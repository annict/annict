# typed: false
# frozen_string_literal: true

RSpec.describe "GET /db/organizations", type: :request do
  it "ユーザーがサインインしていない場合、組織の一覧を表示する" do
    organization = FactoryBot.create(:organization)

    get "/db/organizations"

    expect(response.status).to eq(200)
    expect(response.body).to include(organization.name)
  end

  it "ユーザーがサインインしている場合、組織の一覧を表示する" do
    user = FactoryBot.create(:registered_user)
    organization = FactoryBot.create(:organization)

    login_as(user, scope: :user)
    get "/db/organizations"

    expect(response.status).to eq(200)
    expect(response.body).to include(organization.name)
  end

  it "削除された組織は表示されない" do
    organization = FactoryBot.create(:organization)
    deleted_organization = FactoryBot.create(:organization, deleted_at: Time.current)

    get "/db/organizations"

    expect(response.status).to eq(200)
    expect(response.body).to include(organization.name)
    expect(response.body).not_to include(deleted_organization.name)
  end

  it "組織一覧がID降順で表示される" do
    old_organization = FactoryBot.create(:organization, name: "古い組織")
    new_organization = FactoryBot.create(:organization, name: "新しい組織")

    get "/db/organizations"

    expect(response.status).to eq(200)
    # 新しい組織が古い組織より先に表示される
    expect(response.body.index(new_organization.name)).to be < response.body.index(old_organization.name)
  end

  it "ページパラメータを指定してページネーションが機能する" do
    # 101個の組織を作成（1ページ100件なので、2ページ目に1件表示される）
    101.times do |i|
      FactoryBot.create(:organization, name: "組織#{i}")
    end

    get "/db/organizations", params: {page: 2}

    expect(response.status).to eq(200)
    # 2ページ目なので、最初に作成した組織が表示される
    expect(response.body).to include("組織0")
    # 最後に作成した組織は表示されない（1ページ目に表示されるため）
    expect(response.body).not_to include("組織100")
  end
end
