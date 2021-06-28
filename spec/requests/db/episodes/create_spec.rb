# frozen_string_literal: true

describe "POST /db/works/:work_id/episodes", type: :request do
  context "user does not sign in" do
    let!(:work) { create(:anime) }
    let!(:form_params) do
      {
        rows: "#1,1,The episode"
      }
    end

    it "user can not access this page" do
      post "/db/works/#{work.id}/episodes", params: {db_episode_rows_form: form_params}

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("ログインしてください")

      expect(Episode.all.size).to eq(0)
    end
  end

  context "user who is not editor signs in" do
    let!(:work) { create(:anime) }
    let!(:user) { create(:registered_user) }
    let!(:form_params) do
      {
        rows: "#1,1,The episode"
      }
    end

    before do
      login_as(user, scope: :user)
    end

    it "user can not access" do
      post "/db/works/#{work.id}/episodes", params: {db_episode_rows_form: form_params}

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("アクセスできません")

      expect(Episode.all.size).to eq(0)
    end
  end

  context "user who is editor signs in" do
    let!(:work) { create(:anime) }
    let!(:user) { create(:registered_user, :with_editor_role) }
    let!(:form_params) do
      {
        rows: "第127話,127,逆転！稲妻の戦士\r\n第128話,128,城之内 死す"
      }
    end

    before do
      login_as(user, scope: :user)
    end

    it "user can create episode" do
      expect(Episode.all.size).to eq(0)

      post "/db/works/#{work.id}/episodes", params: {db_episode_rows_form: form_params}

      expect(response.status).to eq(302)
      expect(flash[:notice]).to eq("登録しました")

      expect(Episode.all.size).to eq(2)
      episode = Episode.last

      expect(episode.number).to eq("第128話")
      expect(episode.raw_number).to eq(128)
      expect(episode.title).to eq("城之内 死す")
    end
  end
end
