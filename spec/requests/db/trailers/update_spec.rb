# typed: false
# frozen_string_literal: true

RSpec.describe "PATCH /db/trailers/:id", type: :request do
  it "ログインしていないとき、ログインページにリダイレクトすること" do
    trailer = create(:trailer)
    old_trailer = trailer.attributes
    trailer_params = {
      url: "https://www.youtube.com/watch?v=nGgm5yBznTM",
      title: "タイトル更新",
      sort_number: "200"
    }

    patch "/db/trailers/#{trailer.id}", params: {trailer: trailer_params}
    trailer.reload

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("ログインしてください")
    expect(trailer.title).to eq(old_trailer["title"])
  end

  it "エディター権限がないユーザーがログインしているとき、アクセスできないこと" do
    user = create(:registered_user)
    trailer = create(:trailer)
    old_trailer = trailer.attributes
    trailer_params = {
      url: "https://www.youtube.com/watch?v=nGgm5yBznTM",
      title: "タイトル更新",
      sort_number: "200"
    }

    login_as(user, scope: :user)
    patch "/db/trailers/#{trailer.id}", params: {trailer: trailer_params}
    trailer.reload

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("アクセスできません")
    expect(trailer.title).to eq(old_trailer["title"])
  end

  it "エディター権限があるユーザーがログインしているとき、トレーラーを更新できること" do
    user = create(:registered_user, :with_editor_role)
    trailer = create(:trailer)
    old_trailer = trailer.attributes
    trailer_params = {
      url: "https://www.youtube.com/watch?v=nGgm5yBznTM",
      title: "タイトル更新",
      sort_number: "200"
    }

    login_as(user, scope: :user)
    expect(trailer.title).to eq(old_trailer["title"])

    patch "/db/trailers/#{trailer.id}", params: {trailer: trailer_params}
    trailer.reload

    expect(response.status).to eq(302)
    expect(flash[:notice]).to eq("更新しました")
    expect(trailer.title).to eq("タイトル更新")
    expect(trailer.url).to eq("https://www.youtube.com/watch?v=nGgm5yBznTM")
    expect(trailer.sort_number).to eq(200)
  end

  it "エディター権限があるユーザーがログインしているとき、存在しないトレーラーで404エラーになること" do
    user = create(:registered_user, :with_editor_role)
    trailer_params = {
      url: "https://www.youtube.com/watch?v=nGgm5yBznTM",
      title: "タイトル更新",
      sort_number: "200"
    }

    login_as(user, scope: :user)

    patch "/db/trailers/invalid-id", params: {trailer: trailer_params

    expect(response.status).to eq(404)
  end
end
