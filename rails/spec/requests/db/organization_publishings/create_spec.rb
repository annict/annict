# typed: false
# frozen_string_literal: true

RSpec.describe "POST /db/organizations/:id/publishing", type: :request do
  it "ログインしていないとき、ログインページにリダイレクトすること" do
    organization = create(:organization, :unpublished)

    post "/db/organizations/#{organization.id}/publishing"
    organization.reload

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("ログインしてください")
    expect(organization.published?).to eq(false)
  end

  it "エディター権限がないユーザーがログインしているとき、アクセスできないこと" do
    user = create(:registered_user)
    organization = create(:organization, :unpublished)
    login_as(user, scope: :user)

    post "/db/organizations/#{organization.id}/publishing"
    organization.reload

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("アクセスできません")
    expect(organization.published?).to eq(false)
  end

  it "エディター権限があるユーザーがログインしているとき、組織を公開できること" do
    user = create(:registered_user, :with_editor_role)
    organization = create(:organization, :unpublished)
    login_as(user, scope: :user)

    expect(organization.published?).to eq(false)

    post "/db/organizations/#{organization.id}/publishing"
    organization.reload

    expect(response.status).to eq(302)
    expect(flash[:notice]).to eq("公開しました")
    expect(organization.published?).to eq(true)
  end
end
