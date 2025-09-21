# typed: false
# frozen_string_literal: true

RSpec.describe "GET /people/:person_id/fans", type: :request do
  it "人物のファン一覧が表示されること" do
    person = create(:person)
    user = create(:registered_user)
    create(:person_favorite, user: user, person: person)

    get "/people/#{person.id}/fans"

    expect(response.status).to eq(200)
    expect(response.body).to include(person.name)
    expect(response.body).to include(user.profile.name)
  end

  it "複数のファンがwatched_works_countの降順で表示されること" do
    person = create(:person)
    user1 = create(:registered_user)
    user2 = create(:registered_user)
    user3 = create(:registered_user)
    create(:person_favorite, user: user1, person: person, watched_works_count: 10)
    create(:person_favorite, user: user2, person: person, watched_works_count: 50)
    create(:person_favorite, user: user3, person: person, watched_works_count: 30)

    get "/people/#{person.id}/fans"

    expect(response.status).to eq(200)
    response_body = response.body
    # watched_works_countが多い順に表示されていることを確認
    user2_index = response_body.index(user2.profile.name)
    user3_index = response_body.index(user3.profile.name)
    user1_index = response_body.index(user1.profile.name)
    expect(user2_index).to be < user3_index
    expect(user3_index).to be < user1_index
  end

  it "削除されたユーザーは表示されないこと" do
    person = create(:person)
    active_user = create(:registered_user)
    deleted_user = create(:registered_user, deleted_at: Time.current)
    create(:person_favorite, user: active_user, person: person)
    create(:person_favorite, user: deleted_user, person: person)

    get "/people/#{person.id}/fans"

    expect(response.status).to eq(200)
    expect(response.body).to include(active_user.profile.name)
    expect(response.body).not_to include(deleted_user.profile.name)
  end

  it "削除された人物の場合は404エラーになること" do
    deleted_person = create(:person, deleted_at: Time.current)

    get "/people/#{deleted_person.id}/fans"

    expect(response.status).to eq(404)
  end

  it "存在しない人物の場合は404エラーになること" do
    get "/people/99999999/fans"

    expect(response.status).to eq(404)
  end
end
