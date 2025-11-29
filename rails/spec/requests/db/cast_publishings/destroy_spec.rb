# typed: false
# frozen_string_literal: true

RSpec.describe "DELETE /db/casts/:id/publishing", type: :request do
  it "ログインしていないとき、ログインページにリダイレクトすること" do
    cast = create(:cast, :published)

    delete "/db/casts/#{cast.id}/publishing"
    cast.reload

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("ログインしてください")
    expect(cast.published?).to eq(true)
  end

  it "編集者でないユーザーがログインしているとき、アクセスできないこと" do
    user = create(:registered_user)
    cast = create(:cast, :published)
    login_as(user, scope: :user)

    delete "/db/casts/#{cast.id}/publishing"
    cast.reload

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("アクセスできません")
    expect(cast.published?).to eq(true)
  end

  it "編集者がログインしているとき、キャストを非公開にできること" do
    user = create(:registered_user, :with_editor_role)
    cast = create(:cast, :published)
    login_as(user, scope: :user)

    expect(cast.published?).to eq(true)

    delete "/db/casts/#{cast.id}/publishing"
    cast.reload

    expect(response.status).to eq(302)
    expect(flash[:notice]).to eq("非公開にしました")
    expect(cast.published?).to eq(false)
  end
end
