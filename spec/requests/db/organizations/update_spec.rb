# typed: false
# frozen_string_literal: true

RSpec.describe "PATCH /db/organizations/:id", type: :request do
  it "ログインしていないとき、ログインページにリダイレクトされること" do
    organization = create(:organization)
    old_organization = organization.attributes
    organization_params = {
      name: "御三家"
    }

    patch "/db/organizations/#{organization.id}", params: {organization: organization_params}
    organization.reload

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("ログインしてください")
    expect(organization.name).to eq(old_organization["name"])
  end

  it "エディタ権限を持たないユーザーがログインしているとき、アクセスできないこと" do
    user = create(:registered_user)
    organization = create(:organization)
    old_organization = organization.attributes
    organization_params = {
      name: "御三家"
    }

    login_as(user, scope: :user)

    patch "/db/organizations/#{organization.id}", params: {organization: organization_params}
    organization.reload

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("アクセスできません")
    expect(organization.name).to eq(old_organization["name"])
  end

  it "エディタ権限を持つユーザーがログインしているとき、団体情報を更新できること" do
    user = create(:registered_user, :with_editor_role)
    organization = create(:organization)
    old_organization = organization.attributes
    organization_params = {
      name: "御三家"
    }

    login_as(user, scope: :user)

    expect(organization.name).to eq(old_organization["name"])

    patch "/db/organizations/#{organization.id}", params: {organization: organization_params}
    organization.reload

    expect(response.status).to eq(302)
    expect(flash[:notice]).to eq("更新しました")
    expect(organization.name).to eq("御三家")
  end

  it "エディタ権限を持つユーザーがログインしているとき、複数のフィールドを更新できること" do
    user = create(:registered_user, :with_editor_role)
    organization = create(:organization)
    organization_params = {
      name: "新しい団体名",
      name_en: "New Organization Name",
      name_kana: "アタラシイダンタイメイ",
      url: "https://example.com",
      url_en: "https://example.com/en",
      wikipedia_url: "https://ja.wikipedia.org/wiki/example",
      wikipedia_url_en: "https://en.wikipedia.org/wiki/example",
      twitter_username: "example_org",
      twitter_username_en: "example_org_en"
    }

    login_as(user, scope: :user)

    patch "/db/organizations/#{organization.id}", params: {organization: organization_params}
    organization.reload

    expect(response.status).to eq(302)
    expect(flash[:notice]).to eq("更新しました")
    expect(organization.name).to eq("新しい団体名")
    expect(organization.name_en).to eq("New Organization Name")
    expect(organization.name_kana).to eq("アタラシイダンタイメイ")
    expect(organization.url).to eq("https://example.com")
    expect(organization.url_en).to eq("https://example.com/en")
    expect(organization.wikipedia_url).to eq("https://ja.wikipedia.org/wiki/example")
    expect(organization.wikipedia_url_en).to eq("https://en.wikipedia.org/wiki/example")
    expect(organization.twitter_username).to eq("example_org")
    expect(organization.twitter_username_en).to eq("example_org_en")
  end

  it "エディタ権限を持つユーザーがログインしているとき、無効なデータで更新に失敗すること" do
    user = create(:registered_user, :with_editor_role)
    organization = create(:organization, name: "元の名前")
    organization_params = {
      name: ""
    }

    login_as(user, scope: :user)

    patch "/db/organizations/#{organization.id}", params: {organization: organization_params}
    organization.reload

    expect(response.status).to eq(422)
    expect(organization.name).to eq("元の名前")
  end

  it "エディタ権限を持つユーザーがログインしているとき、削除済みの団体は更新できないこと" do
    user = create(:registered_user, :with_editor_role)
    organization = create(:organization, deleted_at: Time.current)
    organization_params = {
      name: "新しい名前"
    }

    login_as(user, scope: :user)

    patch "/db/organizations/#{organization.id

    expect(response.status).to eq(404)
  end
end
