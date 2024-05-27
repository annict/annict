# typed: false
# frozen_string_literal: true

describe "POST /db/works/:work_id/trailers", type: :request do
  context "user does not sign in" do
    let!(:work) { create(:work) }
    let!(:form_params) do
      {
        rows: "https://www.youtube.com/watch?v=nGgm5yBznTM,第1弾"
      }
    end

    it "user can not access this page" do
      post "/db/works/#{work.id}/trailers", params: {deprecated_db_trailer_rows_form: form_params}

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("ログインしてください")

      expect(Trailer.all.size).to eq(0)
    end
  end

  context "user who is not editor signs in" do
    let!(:work) { create(:work) }
    let!(:user) { create(:registered_user) }
    let!(:form_params) do
      {
        rows: "https://www.youtube.com/watch?v=nGgm5yBznTM,第1弾"
      }
    end

    before do
      login_as(user, scope: :user)
    end

    it "user can not access" do
      post "/db/works/#{work.id}/trailers", params: {deprecated_db_trailer_rows_form: form_params}

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("アクセスできません")

      expect(Trailer.all.size).to eq(0)
    end
  end

  context "user who is editor signs in" do
    let!(:work) { create(:work) }
    let!(:user) { create(:registered_user, :with_editor_role) }
    let!(:form_params) do
      {
        rows: "https://www.youtube.com/watch?v=nGgm5yBznTM,第1弾"
      }
    end

    before do
      login_as(user, scope: :user)
    end

    it "user can create trailer" do
      expect(Trailer.all.size).to eq(0)

      post "/db/works/#{work.id}/trailers", params: {deprecated_db_trailer_rows_form: form_params}

      expect(response.status).to eq(302)
      expect(flash[:notice]).to eq("登録しました")

      expect(Trailer.all.size).to eq(1)
      trailer = Trailer.last

      expect(trailer.title).to eq("第1弾")
    end
  end
end
