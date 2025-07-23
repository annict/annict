# typed: false
# frozen_string_literal: true

RSpec.describe "DELETE /db/organizations/:id/publishing", type: :request do
  it "ログインしていないとき、ログインページにリダイレクトされること" do
    organization = create(:organization, :published)

    delete "/db/organizations/#{organization.id}/publishing"
    organization.reload

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("ログインしてください")
    expect(organization.published?).to eq(true)
  end

  it "編集者権限を持たないユーザーがログインしているとき、アクセスが拒否されること" do
    user = create(:registered_user)
    organization = create(:organization, :published)
    login_as(user, scope: :user)

    delete "/db/organizations/#{organization.id}/publishing"
    organization.reload

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("アクセスできません")
    expect(organization.published?).to eq(true)
  end

  it "編集者権限を持つユーザーがログインしているとき、団体を非公開にできること" do
    user = create(:registered_user, :with_editor_role)
    organization = create(:organization, :published)
    login_as(user, scope: :user)

    expect(organization.published?).to eq(true)

    delete "/db/organizations/#{organization.id}/publishing"
    organization.reload

    expect(response.status).to eq(302)
    expect(flash[:notice]).to eq("非公開にしました")
    expect(organization.published?).to eq(false)
  end

  it "削除済みの団体の場合、404エラーになること" do
    user = create(:registered_user, :with_editor_role)
    organization = create(:organization, :published, :deleted)
    login_as(user, scope: :user)

    expect { delete "/db/organizations/#{organization.id}/publishing" }
      .to raise_error(ActiveRecord::RecordNotFound)
  end

  it "非公開の団体を非公開にしようとした場合、404エラーになること" do
    user = create(:registered_user, :with_editor_role)
    organization = create(:organization, :unpublished)
    login_as(user, scope: :user)

    expect { delete "/db/organizations/#{organization.id}/publishing" }
      .to raise_error(ActiveRecord::RecordNotFound)
  end

  it "存在しない団体IDの場合、404エラーになること" do
    user = create(:registered_user, :with_editor_role)
    login_as(user, scope: :user)

    expect { delete "/db/organizations/invalid-id/publishing" }
      .to raise_error(ActiveRecord::RecordNotFound)
  end
end
