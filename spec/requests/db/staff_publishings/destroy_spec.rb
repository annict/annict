# typed: false
# frozen_string_literal: true

RSpec.describe "DELETE /db/staffs/:id/publishing", type: :request do
  it "ログインしていないとき、ログインページにリダイレクトしページが非公開にならないこと" do
    staff = FactoryBot.create(:staff, :published)

    delete "/db/staffs/#{staff.id}/publishing"
    staff.reload

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("ログインしてください")
    expect(staff.published?).to eq(true)
  end

  it "エディター権限がないユーザーがログインしているとき、アクセス拒否されページが非公開にならないこと" do
    user = FactoryBot.create(:registered_user)
    staff = FactoryBot.create(:staff, :published)

    login_as(user, scope: :user)

    delete "/db/staffs/#{staff.id}/publishing"
    staff.reload

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("アクセスできません")
    expect(staff.published?).to eq(true)
  end

  it "エディター権限があるユーザーがログインしているとき、スタッフ情報を非公開にできること" do
    user = FactoryBot.create(:registered_user, :with_editor_role)
    staff = FactoryBot.create(:staff, :published)

    login_as(user, scope: :user)

    expect(staff.published?).to eq(true)

    delete "/db/staffs/#{staff.id}/publishing"
    staff.reload

    expect(response.status).to eq(302)
    expect(flash[:notice]).to eq("非公開にしました")
    expect(staff.published?).to eq(false)
  end

  it "エディター権限があるユーザーがログインしているとき、存在しないスタッフIDを指定すると404エラーになること" do
    user = FactoryBot.create(:registered_user, :with_editor_role)

    login_as(user, scope: :user)

    delete "/db/staffs/non-existent-id/publishing"

    expect(response).to have_http_status(404)
  end

  it "エディター権限があるユーザーがログインしているとき、すでに非公開のスタッフを指定すると404エラーになること" do
    user = FactoryBot.create(:registered_user, :with_editor_role)
    staff = FactoryBot.create(:staff, :unpublished)

    login_as(user, scope: :user)

    delete "/db/staffs/#{staff.id}/publishing"

    expect(response).to have_http_status(404)
  end
end
