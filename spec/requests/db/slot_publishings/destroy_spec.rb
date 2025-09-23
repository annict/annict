# typed: false
# frozen_string_literal: true

RSpec.describe "DELETE /db/slots/:id/publishing", type: :request do
  it "ログインしていないとき、アクセスできずスロットが公開状態のままであること" do
    slot = create(:slot, :published)

    delete "/db/slots/#{slot.id}/publishing"
    slot.reload

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("ログインしてください")
    expect(slot.published?).to eq(true)
  end

  it "編集者権限がないユーザーでログインしているとき、アクセスできずスロットが公開状態のままであること" do
    user = create(:registered_user)
    slot = create(:slot, :published)

    login_as(user, scope: :user)

    delete "/db/slots/#{slot.id}/publishing"
    slot.reload

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("アクセスできません")
    expect(slot.published?).to eq(true)
  end

  it "編集者権限があるユーザーでログインしているとき、スロットを非公開にできること" do
    user = create(:registered_user, :with_editor_role)
    slot = create(:slot, :published)

    login_as(user, scope: :user)

    expect(slot.published?).to eq(true)

    delete "/db/slots/#{slot.id}/publishing"
    slot.reload

    expect(response.status).to eq(302)
    expect(flash[:notice]).to eq("非公開にしました")
    expect(slot.published?).to eq(false)
  end

  it "編集者権限があるユーザーでログインしているとき、存在しないスロットへのリクエストで404エラーになること" do
    user = create(:registered_user, :with_editor_role)

    login_as(user, scope: :user)

    delete "/db/slots/00000000-0000-0000-0000-000000000000/publishing"

    expect(response).to have_http_status(:not_found)
  end
end
