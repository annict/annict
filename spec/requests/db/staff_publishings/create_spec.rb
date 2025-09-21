# typed: false
# frozen_string_literal: true

RSpec.describe "POST /db/staffs/:id/publishing", type: :request do
  it "ログインしていないとき、ログインページにリダイレクトすること" do
    staff = create(:staff, :unpublished)

    post "/db/staffs/#{staff.id}/publishing"
    staff.reload

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("ログインしてください")
    expect(staff.published?).to eq(false)
  end

  it "エディター権限を持たないユーザーでログインしているとき、アクセスできないこと" do
    user = create(:registered_user)
    staff = create(:staff, :unpublished)
    login_as(user, scope: :user)

    post "/db/staffs/#{staff.id}/publishing"
    staff.reload

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("アクセスできません")
    expect(staff.published?).to eq(false)
  end

  it "エディター権限を持つユーザーでログインしているとき、スタッフを公開できること" do
    user = create(:registered_user, :with_editor_role)
    staff = create(:staff, :unpublished)
    login_as(user, scope: :user)

    expect(staff.published?).to eq(false)

    post "/db/staffs/#{staff.id}/publishing"
    staff.reload

    expect(response.status).to eq(302)
    expect(flash[:notice]).to eq("公開しました")
    expect(staff.published?).to eq(true)
  end

  it "エディター権限を持つユーザーでログインしているとき、すでに公開されているスタッフにアクセスすると404エラーになること" do
    user = create(:registered_user, :with_editor_role)
    staff = create(:staff, :published)
    login_as(user, scope: :user)

    post "/db/staffs/#{staff.id}/publishing"

    expect(response.status).to eq(404)
  end

  it "エディター権限を持つユーザーでログインしているとき、存在しないスタッフIDにアクセスすると404エラーになること" do
    user = create(:registered_user, :with_editor_role)
    login_as(user, scope: :user)

    post "/db/staffs/9999999/publishing"

    expect(response.status).to eq(404)
  end
end
