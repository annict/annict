-- name: ListDBEpisodes :many
-- 直前のエピソードはepisodes.prev_episode_idを読まず、sort_number昇順の隣接行から
-- 導出する。ウィンドウはCTEの中で作品の一覧全体に対して評価され、LIMIT / OFFSETが
-- 1ページに絞り込む前に確定するため、ページ末尾の行も次ページに載るエピソードを指せる。
WITH work_episodes AS (
    SELECT
        e.id,
        e.work_id,
        e.number,
        e.raw_number,
        e.sort_number,
        e.title,
        e.title_ro,
        e.title_en,
        e.episode_records_count,
        e.unpublished_at,
        e.deleted_at,
        LAG(e.id) OVER (ORDER BY e.sort_number, e.id) AS prev_episode_id
    FROM episodes e
    WHERE e.work_id = sqlc.arg('work_id')
        AND e.deleted_at IS NULL
)
SELECT
    we.id,
    we.work_id,
    we.number,
    we.raw_number,
    we.sort_number,
    we.title,
    we.title_ro,
    we.title_en,
    we.episode_records_count,
    we.unpublished_at,
    we.deleted_at,
    prev.number AS prev_number,
    prev.raw_number AS prev_raw_number
FROM work_episodes we
LEFT JOIN episodes prev ON prev.id = we.prev_episode_id
ORDER BY we.sort_number DESC, we.id DESC
LIMIT sqlc.arg('per_page')
OFFSET sqlc.arg('page_offset')::bigint;

-- name: CountDBEpisodes :one
SELECT COUNT(*)
FROM episodes e
WHERE e.work_id = sqlc.arg('work_id')
    AND e.deleted_at IS NULL;

-- name: GetEpisodeForEditByID :one
-- 編集フォームは、エピソードの編集対象カラムと、ページが親作品から必要とする2つの
-- カラム (見出しに使うtitleと、共有の作品サブナビが使うno_episodes) を一緒に読む。
-- 1行で両方を賄うため、フォームを開くのに往復は1回で済む。
--
-- 削除済みエピソードはdeleted_atで除外し (編集アクションが使うRailsの
-- Episode.without_deleted.findと同じ)、作品もエピソード一覧と同じ条件で絞る。作品が
-- 失われたエピソードを、その作品を見出しとサブナビで指すページから編集させないため。
--
-- updated_atはフォームがhiddenで持ち回る版。古い読み取りに対する送信を更新側で
-- 却下できるようにする。
SELECT
    e.id,
    e.work_id,
    e.number,
    e.raw_number,
    e.sort_number,
    e.title,
    e.title_en,
    e.updated_at,
    w.title AS work_title,
    w.no_episodes AS work_no_episodes
FROM episodes e
INNER JOIN works w ON w.id = e.work_id
WHERE e.id = $1
    AND e.deleted_at IS NULL
    AND w.deleted_at IS NULL;

-- name: GetEpisodeForUpdateByID :one
-- 更新が読むのは、送信された値が運ばないものだけ。title_roと状態のタイムスタンプ
-- (animesへの両書きが写像するがフォームでは編集しない) と、両書きを行うか自体を決める2つの
-- マッピングカラム (エピソード自身のanimeと親作品のanime)。編集対象のカラムは送信された値が
-- 置き換えるため読まない。
--
-- 絞り込みはGetEpisodeForEditByIDと揃える。編集フォームに到達できるエピソードと、送信を
-- 受け付けるエピソードを一致させるため。
SELECT
    e.id,
    e.work_id,
    e.title_ro,
    e.unpublished_at,
    e.deleted_at,
    e.anime_id,
    w.anime_id AS parent_anime_id
FROM episodes e
INNER JOIN works w ON w.id = e.work_id
WHERE e.id = $1
    AND e.deleted_at IS NULL
    AND w.deleted_at IS NULL;

-- name: LockWorkForEpisodeUpdateByID :one
-- 隣接エピソードを読む前に、削除されていない親作品をロックする。UpdateDBEpisodeとは
-- 別の文にすることで、ここで待機したトランザクションが、後で移動前後の隣接行を導出するときに
-- READ COMMITTEDの新しいスナップショットを得られるようにする。返したidは、編集フォーム用の
-- 事前読み取り後に対象が別作品へ移った場合の却下にも使う。
--
-- FOR SHAREではなくtouched_workが必要とする強さで最初から取るのは、同一トランザクションの
-- 後段で共有ロックを排他ロックへ昇格させると、同じ作品への2つの送信が互いの共有ロックを
-- 待ち合い、PostgreSQLが片方をデッドロックで中断するためである。
--
-- また、このロックの保持中はこの作品の並び順が固定される。Railsで一覧を並べ替え得る経路は
-- いずれもworks行も書くためここで待たされる (DB管理画面の編集・作成・削除はすべて
-- belongs_to :work, touch: trueかcounter_culture :workを通る)。Goの一括作成も
-- LockWorkForEpisodeCreateByIDで同じロックを取る。したがって隣接行は本ステートメントの後に
-- 導出でき、commitまで有効なままである。
SELECT w.id
FROM episodes e
JOIN works w ON w.id = e.work_id
WHERE e.id = sqlc.arg('id')
    AND e.deleted_at IS NULL
    AND w.deleted_at IS NULL
FOR NO KEY UPDATE OF w;

-- name: ListEpisodeIDsForEpisodeUpdateByID :many
-- UpdateDBEpisodeがこの後に書く行と、隣接行として参照する行をid昇順で列挙する。編集対象
-- 自身・移動前の直前行・移動前の直後行・移動後の直前行・移動後の直後行の5行である。呼び出し側
-- はこれだけをロックするため、1回の編集のロック範囲が作品のエピソード数に比例して増えない。
--
-- この導出が有効なのは、LockWorkForEpisodeUpdateByIDが既に親作品を保持しており、並び順を
-- 変える書き込みがそれを迂回できないため。唯一の例外はAnnict::DataCare::MoveEpisodeで、
-- episodes.work_idをupdate_columnで書くため作品ロックを取らない。本ステートメントの後に
-- この作品へ移されてきた行は列挙もロックもされない。これは手動のデータ整備操作であり、また
-- UpdateDBEpisodeが自身のスナップショットで導出し直すことで開く窓と同じものである。
WITH current_episode AS (
    SELECT e.id, e.work_id, e.sort_number
    FROM episodes e
    WHERE e.id = sqlc.arg('id')
        AND e.work_id = sqlc.arg('work_id')
        AND e.deleted_at IS NULL
), former_preceding_episode AS (
    SELECT p.id
    FROM episodes p
    JOIN current_episode ce ON ce.work_id = p.work_id
    WHERE p.deleted_at IS NULL
        AND p.id <> ce.id
        AND (
            p.sort_number < ce.sort_number
            OR (p.sort_number = ce.sort_number AND p.id < ce.id)
        )
    ORDER BY p.sort_number DESC, p.id DESC
    LIMIT 1
), former_following_episode AS (
    SELECT f.id
    FROM episodes f
    JOIN current_episode ce ON ce.work_id = f.work_id
    WHERE f.deleted_at IS NULL
        AND f.id <> ce.id
        AND (
            f.sort_number > ce.sort_number
            OR (f.sort_number = ce.sort_number AND f.id > ce.id)
        )
    ORDER BY f.sort_number, f.id
    LIMIT 1
), preceding_episode AS (
    SELECT p.id
    FROM episodes p
    JOIN current_episode ce ON ce.work_id = p.work_id
    WHERE p.deleted_at IS NULL
        AND p.id <> ce.id
        AND (
            p.sort_number < sqlc.arg('sort_number')
            OR (p.sort_number = sqlc.arg('sort_number') AND p.id < ce.id)
        )
    ORDER BY p.sort_number DESC, p.id DESC
    LIMIT 1
), following_episode AS (
    SELECT f.id
    FROM episodes f
    JOIN current_episode ce ON ce.work_id = f.work_id
    WHERE f.deleted_at IS NULL
        AND f.id <> ce.id
        AND (
            f.sort_number > sqlc.arg('sort_number')
            OR (f.sort_number = sqlc.arg('sort_number') AND f.id > ce.id)
        )
    ORDER BY f.sort_number, f.id
    LIMIT 1
)
SELECT id FROM current_episode
UNION
SELECT id FROM former_preceding_episode
UNION
SELECT id FROM former_following_episode
UNION
SELECT id FROM preceding_episode
UNION
SELECT id FROM following_episode
ORDER BY id;

-- name: LockEpisodesForEpisodeUpdateByIDs :exec
-- 隣接リンクの導出・書き込み前に、列挙されたエピソードをid昇順でロックする。Railsは
-- エピソードをロックしてから作品をtouchし、Goの更新順序と逆になるため、NOWAITにより循環を
-- 完成させず本トランザクションを中断する。UseCaseはトランザクション全体をrollbackして短時間
-- 後に再試行し、その間にRailsは解放された作品ロックを取得して完了できる。UpdateDBEpisodeが
-- 更新する行と、prev_episode_idから参照する移動先の直前行は、Railsが保持するロックで待たされ
-- 得るため、列挙された隣接行を先取りすれば足りる。idは
-- ListEpisodeIDsForEpisodeUpdateByIDが既に整列済みで、ORDER BYがその順序でのロック取得を保つ。
--
-- 行は読み戻さない。本ステートメントはロック句のためだけに存在する。
SELECT e.id
FROM episodes e
WHERE e.id = ANY(sqlc.arg('ids')::bigint[])
ORDER BY e.id
FOR NO KEY UPDATE OF e NOWAIT;

-- name: UpdateDBEpisode :one
-- 送信が名乗る版の照合は、直前の読み取りとの比較ではなくUPDATEの中で行う。比較と書き込み
-- の間に他の書き込みが挟まらないようにするため。共有カラムはNULL許容のためNULLも明示的な
-- 版であり、IS NOT DISTINCT FROMなら2文に分けずに照合できる。書き込みはupdated_atを進める
-- ので、同じNULL版からの2件目はもう一致しない。したがって1行も返らないことは、その間に
-- 他者が行を書いたことを意味し、呼び出し側は上書きせず競合として報告する。
--
-- prev_episode_idは、エピソード一覧が直前のエピソードを導出するのと同じ規則 (sort_number
-- 昇順、同値ならid) で、送信されたsort_numberから再計算する。Railsはこのカラムを作成時に
-- しか入れず、ずれの修正は編集フォームの選択欄で人が行う前提だったが、Goのフォームにその欄は
-- 無い。公開側のエピソードの前後導線・GraphQL API・REST APIは今もこのカラムを読む。
--
-- 親作品のtouchとepisodes.updateのDB活動の記録は、内容が実際に変わったときにだけ行う。
-- Railsのsave_and_create_activity! も、変更の無い保存では双方を行わないため。比較の対象は
-- 送信された5カラムで、prev_episode_idは含めない。同カラムは入力ではなく並び順から導出される
-- ため、別のエピソードが動いたことによる再計算まで拾うと、共有DBの変更履歴に「その編集者が
-- 行っていない編集」として現れてしまう。データ変更CTEは最後のSELECTが読むかどうかに関わらず
-- 実行されるため、どちらもSELECTにjoinしない。created_activityをjoinすると、何も変わら
-- なかった送信で返すべきidが落ちてしまう。
--
-- 移動は、エピソードの前後にある2行が別の行を隣接として名乗ったままにするため、その2行も
-- 同一文で張り替える。移動前にエピソードの直後だった行は移動前の直前のエピソードを名乗るように
-- し、移動後に直後になる行はエピソード自身を名乗るようにする。2行が同一の場合、エピソードはどの
-- 行も跨いでおらず、どちらも書かない (sort_numberが変わらない送信はすべてこれに該当する)。作品
-- 全体の再計算は意図的に行わない。一括作成はRailsのafter_createコールバックと同じ規則で
-- prev_episode_idを入れており、その値は一覧が導出する隣接行と意図的に異なるため、一括再計算は
-- 無関係な編集でその判断を上書きしてしまう。
--
-- 張り替えはdb_activityを作らず、親作品もtouchせず、隣接行のupdated_atも進めない。並び順
-- から導出される値の維持であって、誰かが行った編集の適用ではないため (Railsが同カラムを
-- update_columnで書くのも同じ理由)。また隣接行の版を進めると、どのフォームも送信しないカラムを
-- 理由に、他の編集者が開いているフォームを競合にしてしまう。
--
-- 呼び出し側は、先行する別ステートメントで削除されていない親作品と、
-- ListEpisodeIDsForEpisodeUpdateByIDが列挙した隣接行を既にロックしている。このプロトコルを
-- 本ステートメントの外に置くことが重要で、作品ロックを待った後にREAD COMMITTEDの新しい
-- スナップショットで本ステートメントを開始し、直前にcommitした更新を移動前後の隣接行の
-- 読み取りへ反映できる。以下の隣接CTEが呼び出し側からidを受け取らずに導出をやり直すのは、
-- 本ステートメントが書く行を、自身のスナップショットが隣接と判断した行に一致させるため。
-- ここでも作品idを条件にすることで、ロック・事前読み取りした親に書き込みを束縛する。
WITH current_episode AS (
    SELECT e.*
    FROM episodes e
    WHERE e.id = sqlc.arg('id')
        AND e.work_id = sqlc.arg('work_id')
        AND e.deleted_at IS NULL
), preceding_episode AS (
    SELECT p.id
    FROM episodes p
    JOIN current_episode ce ON ce.work_id = p.work_id
    WHERE p.deleted_at IS NULL
        AND p.id <> ce.id
        AND (
            p.sort_number < sqlc.arg('sort_number')
            OR (p.sort_number = sqlc.arg('sort_number') AND p.id < ce.id)
        )
    ORDER BY p.sort_number DESC, p.id DESC
    LIMIT 1
), following_episode AS (
    SELECT f.id
    FROM episodes f
    JOIN current_episode ce ON ce.work_id = f.work_id
    WHERE f.deleted_at IS NULL
        AND f.id <> ce.id
        AND (
            f.sort_number > sqlc.arg('sort_number')
            OR (f.sort_number = sqlc.arg('sort_number') AND f.id > ce.id)
        )
    ORDER BY f.sort_number, f.id
    LIMIT 1
), former_preceding_episode AS (
    SELECT p.id
    FROM episodes p
    JOIN current_episode ce ON ce.work_id = p.work_id
    WHERE p.deleted_at IS NULL
        AND p.id <> ce.id
        AND (
            p.sort_number < ce.sort_number
            OR (p.sort_number = ce.sort_number AND p.id < ce.id)
        )
    ORDER BY p.sort_number DESC, p.id DESC
    LIMIT 1
), former_following_episode AS (
    SELECT f.id
    FROM episodes f
    JOIN current_episode ce ON ce.work_id = f.work_id
    WHERE f.deleted_at IS NULL
        AND f.id <> ce.id
        AND (
            f.sort_number > ce.sort_number
            OR (f.sort_number = ce.sort_number AND f.id > ce.id)
        )
    ORDER BY f.sort_number, f.id
    LIMIT 1
), updated_episode AS (
    UPDATE episodes
    SET
        number = sqlc.narg('number'),
        raw_number = sqlc.narg('raw_number'),
        title = sqlc.narg('title'),
        title_en = sqlc.arg('title_en'),
        sort_number = sqlc.arg('sort_number'),
        prev_episode_id = (SELECT pe.id FROM preceding_episode pe),
        updated_at = NOW()
    FROM current_episode ce
    WHERE episodes.id = ce.id
        AND episodes.deleted_at IS NULL
        AND episodes.updated_at IS NOT DISTINCT FROM sqlc.narg('version')::timestamptz
    RETURNING episodes.*
), crossed_neighbours AS (
    SELECT
        ue.id AS episode_id,
        ff.id AS former_following_id,
        fp.id AS former_preceding_id,
        f.id AS following_id
    FROM updated_episode ue
    LEFT JOIN former_following_episode ff ON TRUE
    LEFT JOIN former_preceding_episode fp ON TRUE
    LEFT JOIN following_episode f ON TRUE
    WHERE ff.id IS DISTINCT FROM f.id
), neighbour_relink AS (
    SELECT cn.former_following_id AS id, cn.former_preceding_id AS prev_episode_id
    FROM crossed_neighbours cn
    WHERE cn.former_following_id IS NOT NULL
    UNION ALL
    SELECT cn.following_id AS id, cn.episode_id AS prev_episode_id
    FROM crossed_neighbours cn
    WHERE cn.following_id IS NOT NULL
), relinked_neighbours AS (
    UPDATE episodes
    SET prev_episode_id = nr.prev_episode_id
    FROM neighbour_relink nr
    WHERE episodes.id = nr.id
), episode_change AS (
    SELECT
        ue.id,
        ue.work_id,
        json_build_object('old', row_to_json(ce), 'new', row_to_json(ue)) AS parameters
    FROM updated_episode ue
    JOIN current_episode ce ON ce.id = ue.id
    WHERE (ce.number, ce.raw_number, ce.title, ce.title_en, ce.sort_number)
        IS DISTINCT FROM
        (ue.number, ue.raw_number, ue.title, ue.title_en, ue.sort_number)
), touched_work AS (
    UPDATE works
    SET updated_at = NOW()
    WHERE works.id IN (SELECT ec.work_id FROM episode_change ec)
), created_activity AS (
    INSERT INTO db_activities (
        user_id,
        trackable_id,
        trackable_type,
        action,
        parameters,
        created_at,
        updated_at,
        root_resource_id,
        root_resource_type
    )
    SELECT
        sqlc.arg('user_id'),
        ec.id,
        'Episode',
        'episodes.update',
        ec.parameters,
        NOW(),
        NOW(),
        ec.work_id,
        'Work'
    FROM episode_change ec
)
SELECT ue.id
FROM updated_episode ue;

-- name: GetEpisodeForArchiveByID :one
-- 非公開の確認ページと、それに続く送信は同じ行を読む。画面上でエピソードを名指しする
-- カラム (number / title)、状態のタイムスタンプ、そしてページが親作品から必要とする2つの
-- カラム (見出しに使うtitleと、共有の作品サブナビが使うno_episodes)。送信が使うanimeの
-- 写像は、このトランザクション前の射影ではなく、ArchiveDBEpisodeが実際に更新した行から得る。
--
-- 状態のタイムスタンプは絞り込みに使わず選択する。ページと送信がどの状態を受け付けるかは
-- model.Episode.DerivedStatusを通じて呼び出し側が決めるため。deleted_atで絞るのは、削除済み
-- エピソードが双方の対象外であるためで、GetEpisodeForEditByIDと揃う。DerivedStatusが欠けの
-- 無い行を読めるよう、カラム自体は併せて運ぶ。
--
-- 絞り込みはGetEpisodeForEditByIDと揃える。確認ページに到達できるエピソードと、送信を
-- 受け付けるエピソードを一致させるため。再公開の送信 (UnarchiveDBEpisode) も同じ行を読み、
-- 逆向きの状態条件を当てるため、両方向が同じエピソードに届く。
SELECT
    e.id,
    e.work_id,
    e.number,
    e.title,
    e.unpublished_at,
    e.deleted_at,
    w.title AS work_title,
    w.no_episodes AS work_no_episodes
FROM episodes e
INNER JOIN works w ON w.id = e.work_id
WHERE e.id = $1
    AND e.deleted_at IS NULL
    AND w.deleted_at IS NULL;

-- name: ArchiveDBEpisode :one
-- 非公開は、エピソードが今も公開中であることを条件とする。古い確認ページからの送信は
-- unpublished_atを再スタンプせず、1行も返さない。作品idは確認ページが前提とした親作品との
-- 一致を要求する。その間にAnnict::DataCare::MoveEpisodeで別作品へ移されたエピソードが、
-- もう所属していない作品のカウンターを減算しないようにするため。親作品も未削除のままであること
-- を要求する。確認ページ用の射影は削除済み作品を除外するが、作品のライフサイクルは本ステート
-- メントの開始前や、別のトランザクションがepisode行をロックしている間にも変わり得るため。
--
-- 更新した行は現在のanime_idを返す。両書きが、実際に非公開にした時点の写像を対象にするため。
-- 確認ページの読み取り後に写像が変わっても、送信が以前のanimeを更新してはならない。
--
-- Railsの非公開はEpisode#update(unpublished_at:) であり、db_activityは作らないが、
-- updated_atを進め、belongs_to :work, touch: trueで親作品をtouchし、counter_cultureで
-- works.episodes_countを減算する (column_nameのラムダは公開中のエピソードだけを数える)。
-- 3つとも本ステートメントで再現する。1文にまとめるのは、カウンターが数える対象の状態から
-- ずれないようにするため (減算は上のUPDATEが実際に行った遷移に対してだけ走る)。最終結果は
-- 作品の更新成功にも依存させる。同時の論理削除によって作品の更新がスキップされた場合、:oneは
-- 行を返さず、呼び出し元が最初のCTEによるepisodeの更新も含めてトランザクションをロール
-- バックする。
--
-- 書き込みはepisodesが先、worksが後で、Railsが保存する順序と同じ。これにより2つの
-- アプリケーションは同じ作品に対してデッドロックせず、順に待ち合う。Goのエピソード更新は
-- 逆順で取る (LockWorkForEpisodeUpdateByIDで作品が先) が、この反転が循環にならないのは、
-- 更新側が必要なエピソードをNOWAITで先取りし、待たずに試行を中断するため。一括作成も作品を
-- 先に取るが、行を挿入するだけのため本ステートメントが保持するエピソードを待つことはない。
WITH archived_episode AS (
    UPDATE episodes
    SET
        unpublished_at = NOW(),
        updated_at = NOW()
    WHERE episodes.id = sqlc.arg('id')
        AND episodes.work_id = sqlc.arg('work_id')
        AND episodes.unpublished_at IS NULL
        AND episodes.deleted_at IS NULL
        AND EXISTS (
            SELECT 1
            FROM works
            WHERE works.id = episodes.work_id
                AND works.deleted_at IS NULL
        )
    RETURNING episodes.id, episodes.work_id, episodes.anime_id
), touched_work AS (
    UPDATE works
    SET
        episodes_count = works.episodes_count - 1,
        updated_at = NOW()
    WHERE works.id IN (SELECT ae.work_id FROM archived_episode ae)
        AND works.deleted_at IS NULL
    RETURNING works.id
)
SELECT ae.id, ae.anime_id
FROM archived_episode ae
INNER JOIN touched_work tw ON tw.id = ae.work_id;

-- name: UnarchiveDBEpisode :one
-- 再公開はArchiveDBEpisodeの逆で、あちらが打つタイムスタンプをクリアし、あちらが引く
-- カウンターを戻す。ガードはすべて同じで、状態の条件だけを逆向きに読む。エピソードが今も非公開
-- であることを条件とする。他者が再公開する前に開いた一覧から、その再公開後に送信された場合は、
-- すでにNULLのunpublished_atをクリアせず1行も返さない。作品idは一覧が名指しした親作品との
-- 一致を要求する。その間にAnnict::DataCare::MoveEpisodeで別作品へ移されたエピソードが、
-- もう所属していない作品の
-- カウンターを加算しないようにするため。親作品も未削除のままであることを要求する。トランザクション
-- 前の射影は削除済み作品を除外するが、作品のライフサイクルは本ステートメントの開始前や、別の
-- トランザクションがepisode行をロックしている間にも変わり得るため。
--
-- 更新した行は現在のanime_idを返す。両書きが、トランザクション前の射影が観測した写像では
-- なく、実際に再公開した時点の写像を対象にするため。
--
-- Railsの再公開はEpisode#update(unpublished_at: nil) であり、db_activityは作らないが、
-- updated_atを進め、belongs_to :work, touch: trueで親作品をtouchし、counter_cultureで
-- works.episodes_countを加算する (column_nameのラムダは公開中のエピソードだけを数えるため、
-- unpublished_atのクリアで行が数え直される)。3つとも本ステートメントで再現する。1文に
-- まとめるのは、カウンターが数える対象の状態からずれないようにするため (加算は上のUPDATEが
-- 実際に行った遷移に対してだけ走る)。最終結果は作品の更新成功にも依存させる。同時の論理削除に
-- よって作品の更新がスキップされた場合、:oneは行を返さず、呼び出し元が最初のCTEによる
-- episodeの更新も含めてトランザクションをロールバックする。
--
-- 書き込みがepisodes先・works後なのはArchiveDBEpisodeが述べる理由による。
WITH published_episode AS (
    UPDATE episodes
    SET
        unpublished_at = NULL,
        updated_at = NOW()
    WHERE episodes.id = sqlc.arg('id')
        AND episodes.work_id = sqlc.arg('work_id')
        AND episodes.unpublished_at IS NOT NULL
        AND episodes.deleted_at IS NULL
        AND EXISTS (
            SELECT 1
            FROM works
            WHERE works.id = episodes.work_id
                AND works.deleted_at IS NULL
        )
    RETURNING episodes.id, episodes.work_id, episodes.anime_id
), touched_work AS (
    UPDATE works
    SET
        episodes_count = works.episodes_count + 1,
        updated_at = NOW()
    WHERE works.id IN (SELECT pe.work_id FROM published_episode pe)
        AND works.deleted_at IS NULL
    RETURNING works.id
)
SELECT pe.id, pe.anime_id
FROM published_episode pe
INNER JOIN touched_work tw ON tw.id = pe.work_id;

-- name: GetEpisodeForDeleteByID :one
-- 削除の送信はページではなく確認アラートから来るため、この射影は書き込みが必要とするもの
-- だけを運ぶ。対象のidと、送信が名指しするとみなす親作品 (DeleteDBEpisodeをその作品に束縛し、
-- 削除の成功時に着地する先でもある)。表示のための読み取りは無いため、エピソードを名指しする
-- カラムは運ばない。
--
-- 削除済みエピソードと、削除済み作品のエピソードは除外する。Db::EpisodesController#destroyの
-- RailsのEpisode.without_deleted.findと、非公開エンドポイント群が届く範囲に揃えるため。
-- 状態はそれ以上絞らない。公開中のエピソードも非公開のエピソードも削除できるため、非公開
-- エンドポイントと違い呼び出し側が当てる状態条件が無く、タイムスタンプを運ぶ必要も無い。
SELECT
    e.id,
    e.work_id
FROM episodes e
INNER JOIN works w ON w.id = e.work_id
WHERE e.id = $1
    AND e.deleted_at IS NULL
    AND w.deleted_at IS NULL;

-- name: DeleteDBEpisode :one
-- 削除は、エピソードがまだ削除されていないことを条件とする。他者が削除する前に開いた一覧から、
-- その削除後に送信された場合は、deleted_atを再スタンプせず、カウンターを2度目に減算せず、
-- 1行も返さない。作品idは一覧が名指しした親作品との一致を要求する。その間に
-- Annict::DataCare::MoveEpisodeで別作品へ移されたエピソードが、もう所属していない作品の
-- カウンターを減算しないようにするため。親作品も
-- 未削除のままであることを要求する。トランザクション前の射影は削除済み作品を除外するが、作品の
-- ライフサイクルは本ステートメントの開始前や、別のトランザクションがepisode行をロックしている
-- 間にも変わり得るため。
--
-- 更新した行は現在のanime_idを返す。両書きが、トランザクション前の射影が観測した写像ではなく、
-- 実際に削除した時点の写像を対象にするため。
--
-- Railsの削除はdestroy_in_batchesで行を消すため、counter_cultureが
-- works.episodes_countを減算し (column_nameのラムダは公開中のエピソードだけを数えるため、
-- すでに非公開のエピソードの削除では何も減らない)、belongs_to :work, touch: trueが
-- works.updated_atを進める。Goの削除はソフトデリートで行が残り、どちらの副作用も自動的には
-- 従わないため、2つとも本ステートメントで再現する。1文にまとめるのは、カウンターが数える対象の
-- 状態からずれないようにするため (減算は上のUPDATEが実際に行った遷移に対してだけ、かつまだ
-- 数えられていた行に対してだけ走る)。最終結果は作品の更新成功にも依存させる。同時の論理削除に
-- よって作品の更新がスキップされた場合、:oneは行を返さず、呼び出し元が最初のCTEによる
-- episodeの更新も含めてトランザクションをロールバックする。
--
-- 書き込みがepisodes先・works後なのはArchiveDBEpisodeが述べる理由による。非公開と違い、
-- ロックする範囲には削除するエピソードを名乗る行も含まれ、下のCTEがそれをエピソードと作品の
-- 間で書く。同一作品に対する2つの削除はprev_episode_idの順に行を取るが、これは循環ではなく
-- 連鎖であり、作品はどちらも最後に取る。Goのエピソード更新はここでも作品を先に取り、必要な
-- エピソードをNOWAITで先取りして、待たずに試行を中断するため、こちらの反転も循環にならない。
--
-- 削除するエピソードを直前として名乗る行は、そのポインタをクリアする。Railsの
-- before_destroy :unset_prev_episode_idが行が消える前に行うことと同じ。Goでは行が残り、
-- prev_episode_id自体にスコープは無い。公開側のエピソードの前後導線・GraphQLの
-- EpisodeType.prev_episode・REST APIはこのカラムが名指しするものをそのまま読むため、残した
-- ポインタは削除済みエピソードを見せてしまう。Railsのコールバックが見つける「公開中の1行」で
-- はなく、名乗っている未削除の行すべてをクリアするのは、非公開の後続行が古いポインタを持ったまま
-- 再公開された瞬間に、それが表に戻るのを防ぐため。後続行を削除するエピソード自身の直前行へ張り
-- 替えることはしない。それはGoの削除後の状態を、後続行に直前を持たせないRailsの削除後の状態
-- と食い違わせるため。
--
-- クリアが書くのはprev_episode_idだけで、変更履歴も親のtouchも後続行の版の更新も行わない。
-- 理由はUpdateDBEpisodeの張り替えが述べるものと同じ。削除するエピソード自身の
-- prev_episode_idはそのままにする。削除済みの行は何も表示しないため。
WITH deleted_episode AS (
    UPDATE episodes
    SET
        deleted_at = NOW(),
        updated_at = NOW()
    WHERE episodes.id = sqlc.arg('id')
        AND episodes.work_id = sqlc.arg('work_id')
        AND episodes.deleted_at IS NULL
        AND EXISTS (
            SELECT 1
            FROM works
            WHERE works.id = episodes.work_id
                AND works.deleted_at IS NULL
        )
    RETURNING episodes.id, episodes.work_id, episodes.anime_id, episodes.unpublished_at
), unlinked_followers AS (
    UPDATE episodes
    SET prev_episode_id = NULL
    FROM deleted_episode de
    WHERE episodes.prev_episode_id = de.id
        AND episodes.id <> de.id
        AND episodes.deleted_at IS NULL
), touched_work AS (
    UPDATE works
    SET
        episodes_count = works.episodes_count - (
            CASE WHEN de.unpublished_at IS NULL THEN 1 ELSE 0 END
        ),
        updated_at = NOW()
    FROM deleted_episode de
    WHERE works.id = de.work_id
        AND works.deleted_at IS NULL
    RETURNING works.id
)
SELECT de.id, de.anime_id
FROM deleted_episode de
INNER JOIN touched_work tw ON tw.id = de.work_id;

-- name: ListEpisodesForAnimeSyncByIDs :many
SELECT
    e.id,
    e.work_id,
    e.title,
    e.title_ro,
    e.title_en,
    e.number,
    e.sort_number,
    e.raw_number,
    e.unpublished_at,
    e.deleted_at,
    e.anime_id,
    w.anime_id AS parent_anime_id
FROM episodes e
JOIN works w ON e.work_id = w.id
WHERE e.id = ANY($1::bigint[])
ORDER BY e.id;

-- name: ListEpisodeIDsAfter :many
SELECT id
FROM episodes
WHERE id > sqlc.arg('after_id')
ORDER BY id
LIMIT sqlc.arg('batch_size');

-- name: UpdateEpisodeAnimeID :exec
UPDATE episodes
SET anime_id = $2
WHERE id = $1;

-- name: CreateEpisode :one
-- anime_idは後から書き戻さず行と一緒に書く。一括作成はエピソード本体より先にその
-- animeを挿入するため、マッピングカラムの値が既に分かっているため。prev_episode_idには
-- 挿入時点でsort_numberが最大のエピソードを入れる (Railsのafter_createコールバックが
-- 入れるのと同じ値)。Annict DBの一覧は直前のエピソードをsort_number順から導出するが、
-- 公開側のエピソードページとGraphQL APIは今もこのカラムを読む。データ変更CTEはさらに、
-- Railsのsave_and_create_activity! と同じepisodes.createのDB活動を、挿入行を
-- parameters.newとして記録する。
WITH created_episode AS (
    INSERT INTO episodes (
        work_id,
        number,
        raw_number,
        title,
        sort_number,
        prev_episode_id,
        anime_id,
        created_at,
        updated_at
    ) VALUES (
        sqlc.arg('work_id'),
        sqlc.narg('number'),
        sqlc.narg('raw_number'),
        sqlc.narg('title'),
        sqlc.arg('sort_number'),
        sqlc.narg('prev_episode_id'),
        sqlc.narg('anime_id'),
        NOW(),
        NOW()
    )
    RETURNING *
), created_activity AS (
    INSERT INTO db_activities (
        user_id,
        trackable_id,
        trackable_type,
        action,
        parameters,
        created_at,
        updated_at,
        root_resource_id,
        root_resource_type
    )
    SELECT
        sqlc.arg('user_id'),
        ce.id,
        'Episode',
        'episodes.create',
        json_build_object('new', row_to_json(ce)),
        NOW(),
        NOW(),
        ce.work_id,
        'Work'
    FROM created_episode ce
    RETURNING id
)
SELECT ce.id
FROM created_episode ce
CROSS JOIN created_activity ca;
