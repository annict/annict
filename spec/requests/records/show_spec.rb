# frozen_string_literal: true

describe "GET /@:username/records/:record_id", type: :request do
  let(:user) { create(:registered_user) }

  context "ログインしているとき" do
    before do
      login_as(user, scope: :user)
    end

    context "アニメへの記録を参照するとき" do
      let!(:record) { create(:record, user: user) }
      let!(:anime_record) { create(:work_record, user: user, record: record, body: "最高") }
      let(:anime) { anime_record.work }

      it "記録が表示されること" do
        get "/@#{user.username}/records/#{record.id}"

        expect(response.status).to eq(200)
        expect(response.body).to include(user.profile.name)
        expect(response.body).to include(anime.title)
        expect(response.body).to include("最高")
      end
    end

    context "エピソードへの記録を参照するとき" do
      let!(:episode_record) { create(:episode_record, user: user, body: "楽しかった") }
      let(:record) { episode_record.record }
      let(:anime) { episode_record.work }
      let(:episode) { episode_record.episode }

      it "記録が表示されること" do
        get "/@#{user.username}/records/#{record.id}"

        expect(response.status).to eq(200)
        expect(response.body).to include(user.profile.name)
        expect(response.body).to include(anime.title)
        expect(response.body).to include(episode.number)
        expect(response.body).to include("楽しかった")
      end
    end
  end

  context "ログインしていないとき" do
    context "アニメへの記録を参照したとき" do
      let!(:record) { create(:record, user: user) }
      let!(:anime_record) { create(:work_record, user: user, record: record, body: "最高") }
      let(:anime) { anime_record.work }

      it "記録が表示されること" do
        get "/@#{user.username}/records/#{record.id}"

        expect(response.status).to eq(200)
        expect(response.body).to include(user.profile.name)
        expect(response.body).to include(anime.title)
        expect(response.body).to include("最高")
      end
    end

    context "エピソードへの記録を参照したとき" do
      let!(:episode_record) { create(:episode_record, user: user, body: "楽しかった") }
      let(:record) { episode_record.record }
      let(:anime) { episode_record.work }
      let(:episode) { episode_record.episode }

      it "記録が表示されること" do
        get "/@#{user.username}/records/#{record.id}"

        expect(response.status).to eq(200)
        expect(response.body).to include(user.profile.name)
        expect(response.body).to include(anime.title)
        expect(response.body).to include(episode.number)
        expect(response.body).to include("楽しかった")
      end
    end
  end
end
