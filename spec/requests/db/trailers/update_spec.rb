# frozen_string_literal: true

describe "PATCH /db/trailers/:id", type: :request do
  context "user does not sign in" do
    let!(:trailer) { create(:trailer) }
    let!(:old_trailer) { trailer.attributes }
    let!(:trailer_params) do
      {
        url: "https://www.youtube.com/watch?v=nGgm5yBznTM",
        title: "タイトル更新",
        sort_number: "200"
      }
    end

    it "user can not access this page" do
      patch "/db/trailers/#{trailer.id}", params: {trailer: trailer_params}
      trailer.reload

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("ログインしてください")

      expect(trailer.title).to eq(old_trailer["title"])
    end
  end

  context "user who is not editor signs in" do
    let!(:user) { create(:registered_user) }
    let!(:trailer) { create(:trailer) }
    let!(:old_trailer) { trailer.attributes }
    let!(:trailer_params) do
      {
        url: "https://www.youtube.com/watch?v=nGgm5yBznTM",
        title: "タイトル更新",
        sort_number: "200"
      }
    end

    before do
      login_as(user, scope: :user)
    end

    it "user can not access" do
      patch "/db/trailers/#{trailer.id}", params: {trailer: trailer_params}
      trailer.reload

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("アクセスできません")

      expect(trailer.title).to eq(old_trailer["title"])
    end
  end

  context "user who is editor signs in" do
    let!(:user) { create(:registered_user, :with_editor_role) }
    let!(:trailer) { create(:trailer) }
    let!(:old_trailer) { trailer.attributes }
    let!(:trailer_params) do
      {
        url: "https://www.youtube.com/watch?v=nGgm5yBznTM",
        title: "タイトル更新",
        sort_number: "200"
      }
    end
    let!(:attr_names) { trailer_params.keys }

    before do
      login_as(user, scope: :user)
    end

    it "user can update trailer" do
      expect(trailer.title).to eq(old_trailer["title"])

      patch "/db/trailers/#{trailer.id}", params: {trailer: trailer_params}
      trailer.reload

      expect(response.status).to eq(302)
      expect(flash[:notice]).to eq("更新しました")

      expect(trailer.title).to eq("タイトル更新")
    end
  end
end
