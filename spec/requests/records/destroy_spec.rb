# frozen_string_literal: true

describe "DELETE /@:username/records/:record_id", type: :request do
  let(:user) { create :registered_user }
  let!(:anime) { create(:anime) }
  let(:episode) { create(:episode, anime: anime) }
  let!(:record) { create(:record, user: user, anime: anime) }
  let(:episode_record) { create(:episode_record, user: user, record: record, anime: anime, episode: episode) }
  let(:activity_group) { create(:activity_group, user: user, itemable_type: "EpisodeRecord") }
  let!(:activity) { create(:activity, user: user, activity_group: activity_group, itemable: episode_record) }
  let!(:library_entry) { create(:library_entry, user: user, anime: anime, watched_episode_ids: [episode.id]) }

  context "ログインしているとき" do
    before do
      login_as(user, scope: :user)
    end

    it "記録が削除できること" do
      expect(ActivityGroup.count).to eq 1
      expect(Activity.count).to eq 1
      expect(Record.count).to eq 1
      expect(EpisodeRecord.count).to eq 1
      expect(LibraryEntry.count).to eq 1
      expect(library_entry.watched_episode_ids).to eq [episode.id]

      delete "/@#{user.username}/records/#{record.id}"

      expect(ActivityGroup.count).to eq 0
      expect(Activity.count).to eq 0
      expect(Record.count).to eq 0
      expect(EpisodeRecord.count).to eq 0
      expect(LibraryEntry.count).to eq 1
      expect(library_entry.reload.watched_episode_ids).to eq []

      expect(response.status).to eq(302)
      expect(response).to redirect_to("/works/#{anime.id}/episodes/#{episode.id}")
    end
  end
end
