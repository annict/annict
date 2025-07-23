# typed: false
# frozen_string_literal: true

RSpec.describe "GET /forum/posts/:post_id/comments/:comment_id/edit", type: :request do
  it "ログインしていないとき、ログインページにリダイレクトすること" do
    forum_post = create(:forum_post)
    forum_comment = create(:forum_comment, forum_post: forum_post)

    get "/forum/posts/#{forum_post.id}/comments/#{forum_comment.id}/edit"

    expect(response).to redirect_to(new_user_session_path)
    expect(flash[:alert]).to eq("ログインしてください")
  end

  it "ログインしているが、他のユーザーのコメントを編集しようとしたとき、403エラーになること" do
    user = create(:registered_user)
    other_user = create(:registered_user)
    forum_post = create(:forum_post)
    forum_comment = create(:forum_comment, forum_post: forum_post, user: other_user)
    login_as(user, scope: :user)

    expect {
      get "/forum/posts/#{forum_post.id}/comments/#{forum_comment.id}/edit"
    }.to raise_error(Pundit::NotAuthorizedError)
  end

  it "ログインしていて、自分のコメントを編集しようとしたとき、編集画面が表示されること" do
    user = create(:registered_user)
    forum_post = create(:forum_post)
    forum_comment = create(:forum_comment, forum_post: forum_post, user: user, body: "元のコメント内容")
    login_as(user, scope: :user)

    get "/forum/posts/#{forum_post.id}/comments/#{forum_comment.id}/edit"

    expect(response.status).to eq(200)
    expect(response.body).to include("元のコメント内容")
  end

  it "ログインしているとき、存在しない投稿IDが指定された場合、404エラーになること" do
    user = create(:registered_user)
    forum_comment = create(:forum_comment, user: user)
    login_as(user, scope: :user)

    expect {
      get "/forum/posts/99999/comments/#{forum_comment.id}/edit"
    }.to raise_error(ActiveRecord::RecordNotFound)
  end

  it "ログインしているとき、存在しないコメントIDが指定された場合、404エラーになること" do
    user = create(:registered_user)
    forum_post = create(:forum_post)
    login_as(user, scope: :user)

    expect {
      get "/forum/posts/#{forum_post.id}/comments/99999/edit"
    }.to raise_error(ActiveRecord::RecordNotFound)
  end

  it "ログインしているとき、投稿とコメントの関連が正しくない場合、404エラーになること" do
    user = create(:registered_user)
    forum_post1 = create(:forum_post)
    forum_post2 = create(:forum_post)
    forum_comment = create(:forum_comment, forum_post: forum_post1, user: user)
    login_as(user, scope: :user)

    expect {
      get "/forum/posts/#{forum_post2.id}/comments/#{forum_comment.id}/edit"
    }.to raise_error(ActiveRecord::RecordNotFound)
  end
end
