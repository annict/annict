# typed: false
# frozen_string_literal: true

RSpec.describe "DELETE /@:username/records/:record_id", type: :request do
  it "ログインしているとき、記録が削除できること" do
    user = create(:registered_user)
    work = create(:work)
    episode = create(:episode, work: work)
    record = create(:record, user: user, work: work)
    episode_record = create(:episode_record, user: user, record: record, work: work, episode: episode)
    activity_group = create(:activity_group, user: user, itemable_type: "EpisodeRecord")
    create(:activity, user: user, activity_group: activity_group, itemable: episode_record)
    library_entry = create(:library_entry, user: user, work: work, watched_episode_ids: [episode.id])

    login_as(user, scope: :user)

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
    expect(response).to redirect_to("/works/#{work.id}/episodes/#{episode.id}")
  end

  it "ログインしていないとき、ログインページにリダイレクトされること" do
    user = create(:registered_user)
    work = create(:work)
    record = create(:record, user: user, work: work)

    delete "/@#{user.username}/records/#{record.id}"

    expect(response.status).to eq(302)
    expect(response).to redirect_to(new_user_session_path)
  end

  it "他のユーザーの記録を削除しようとしたとき、認可エラーになること" do
    user = create(:registered_user)
    other_user = create(:registered_user)
    work = create(:work)
    record = create(:record, user: other_user, work: work)

    login_as(user, scope: :user)

    expect {
      delete "/@#{other_user.username}/records/#{record.id}"
    }.to raise_error(Pundit::NotAuthorizedError)
  end

  it "存在しないレコードを削除しようとしたとき、404エラーが返されること" do
    user = create(:registered_user)

    login_as(user, scope: :user)

    delete "/@#{user.username}/records/999999"

    expect(response).to have_http_status(:not_found)
  end

  it "存在しないユーザーのレコードを削除しようとしたとき、404エラーが返されること" do
    user = create(:registered_user)

    login_as(user, scope: :user)

    delete "/@nonexistent_user/records/999999"

    expect(response).to have_http_status(:not_found)
  end
end
