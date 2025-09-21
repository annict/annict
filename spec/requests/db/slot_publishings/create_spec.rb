# typed: false
# frozen_string_literal: true

RSpec.describe "POST /db/slots/:id/publishing", type: :request do
  it "ログインしていないとき、アクセスできないこと" do
    slot = create(:slot, :unpublished)

    post "/db/slots/#{slot.id}/publishing"
    slot.reload

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("ログインしてください")
    expect(slot.published?).to eq(false)
  end

  it "編集者権限のないユーザーがログインしているとき、アクセスできないこと" do
    user = create(:registered_user)
    slot = create(:slot, :unpublished)
    login_as(user, scope: :user)

    post "/db/slots/#{slot.id}/publishing"
    slot.reload

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("アクセスできません")
    expect(slot.published?).to eq(false)
  end

  it "編集者権限のあるユーザーがログインしているとき、スロットを公開できること" do
    user = create(:registered_user, :with_editor_role)
    slot = create(:slot, :unpublished)
    login_as(user, scope: :user)

    expect(slot.published?).to eq(false)

    post "/db/slots/#{slot.id}/publishing"
    slot.reload

    expect(response.status).to eq(302)
    expect(flash[:notice]).to eq("公開しました")
    expect(slot.published?).to eq(true)
  end

  it "編集者権限のあるユーザーがログインしているとき、存在しないスロットIDに対してリクエストした場合、エラーになること" do
    user = create(:registered_user, :with_editor_role)
    login_as(user, scope: :user)

    post "/db/slots/invalid-id/publishing"

    expect(response.status).to eq(404)
  end

  it "編集者権限のあるユーザーがログインしているとき、既に公開済みのスロットに対してリクエストした場合、エラーになること" do
    user = create(:registered_user, :with_editor_role)
    slot = create(:slot, :published)
    login_as(user, scope: :user)

    expect {
      post "/db/slots/#{slot.id}/publishing"
    }.to raise_error(ActiveRecord::RecordNotFound)
  end
end
