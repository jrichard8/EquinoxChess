import tempfile
import unittest
from pathlib import Path

import chess

import analyze


class ScoreParsingTests(unittest.TestCase):
    def test_parses_cp_and_mate_signs(self):
        self.assertEqual(analyze.parse_score("+1.48/4 0.01s", 3000).cp, 148)
        self.assertEqual(analyze.parse_score("-0.03/12", 3000).cp, -3)
        positive_mate = analyze.parse_score("+M3/20", 3000)
        negative_mate = analyze.parse_score("-M2/20", 3000)
        self.assertTrue(positive_mate.is_mate)
        self.assertIsNone(positive_mate.cp)
        self.assertEqual(positive_mate.mate_sign, 1)
        self.assertEqual(negative_mate.mate_sign, -1)
        self.assertIsNone(analyze.parse_score("0.050s", 3000))


class KingZoneFeatureTests(unittest.TestCase):
    def board_with_white_king(self):
        board = chess.Board.empty()
        board.set_piece_at(chess.E1, chess.Piece(chess.KING, chess.WHITE))
        board.set_piece_at(chess.A8, chess.Piece(chess.KING, chess.BLACK))
        return board

    def test_slider_includes_first_own_or_enemy_blocker_then_stops(self):
        for blocker_color in (chess.WHITE, chess.BLACK):
            with self.subTest(blocker_color=blocker_color):
                board = self.board_with_white_king()
                board.set_piece_at(chess.E8, chess.Piece(chess.ROOK, chess.BLACK))
                board.set_piece_at(chess.E2, chess.Piece(chess.PAWN, blocker_color))
                features = analyze.king_zone_features(board, chess.WHITE)
                self.assertEqual(features["rook_attackers"], 1)
                self.assertEqual(features["rook_hit_sum"], 1)

        board = self.board_with_white_king()
        board.set_piece_at(chess.E8, chess.Piece(chess.ROOK, chess.BLACK))
        board.set_piece_at(chess.E3, chess.Piece(chess.PAWN, chess.WHITE))
        self.assertEqual(analyze.king_zone_features(board, chess.WHITE)["rook_attackers"], 0)

    def test_one_queen_counts_once_with_multiple_direction_hits(self):
        board = self.board_with_white_king()
        board.set_piece_at(chess.E3, chess.Piece(chess.QUEEN, chess.BLACK))
        features = analyze.king_zone_features(board, chess.WHITE)
        self.assertEqual(features["queen_attackers"], 1)
        self.assertEqual(features["queen_hit_sum"], 3)
        self.assertEqual(features["queen_union_hits"], 3)

    def test_corner_ring_excludes_king_square(self):
        board = chess.Board.empty()
        board.set_piece_at(chess.A1, chess.Piece(chess.KING, chess.WHITE))
        self.assertEqual(
            analyze.king_ring(board, chess.WHITE),
            chess.BB_B1 | chess.BB_A2 | chess.BB_B2,
        )

    def test_game09_golden_pre_and_post(self):
        board = chess.Board("8/8/RP1B4/3b4/8/P3k2p/3r4/6K1 w - - 15 64")
        pre = analyze.king_zone_features(board, chess.WHITE)
        self.assertEqual(
            {key: pre[key] for key in (
                "bishop_attackers", "bishop_hit_sum", "rook_attackers", "rook_hit_sum",
                "queen_attackers", "total_attackers", "total_hit_sum", "total_union_hits",
                "king_in_check",
            )},
            {
                "bishop_attackers": 1, "bishop_hit_sum": 2,
                "rook_attackers": 1, "rook_hit_sum": 3,
                "queen_attackers": 0, "total_attackers": 2,
                "total_hit_sum": 5, "total_union_hits": 4,
                "king_in_check": False,
            },
        )
        board.push_uci("a6a7")
        self.assertEqual(analyze.king_zone_features(board, chess.WHITE), pre)


class ReplayTests(unittest.TestCase):
    PGN = """[Event \"synthetic\"]
[White \"EquinoxChess\"]
[Black \"Opponent\"]
[Result \"*\"]

1. e4 {+2.00/4} e5 {+1.50/8} *

[Event \"synthetic\"]
[White \"Opponent\"]
[Black \"EquinoxChess\"]
[Result \"*\"]

1. d4 {+0.20/8} d5 {+0.50/4} 2. c4 {+M3/8} *
"""

    def test_replay_keeps_mover_pov_gap_mates_and_exact_canary(self):
        with tempfile.TemporaryDirectory() as directory:
            path = Path(directory) / "synthetic.pgn"
            path.write_text(self.PGN, encoding="utf-8")
            start_fen = chess.Board().fen()
            rows, counts = analyze.analyze_pgn(
                path, {start_fen: "start-exact"}, "EquinoxChess", 300, 3000
            )
        self.assertEqual(len(rows), 2)
        self.assertEqual(rows[0]["engine_color"], "white")
        self.assertEqual(rows[0]["loss_swing_cp"], 350)
        self.assertEqual(rows[0]["optimism_gap_cp"], 350)
        self.assertTrue(rows[0]["severe"])
        self.assertEqual(rows[0]["canary_name"], "start-exact")
        self.assertEqual(rows[1]["engine_color"], "black")
        self.assertIsNone(rows[1]["loss_swing_cp"])
        self.assertTrue(rows[1]["loss_mate"])
        self.assertTrue(rows[1]["severe"])
        self.assertEqual(counts["engine_games_white"], 1)
        self.assertEqual(counts["engine_games_black"], 1)


class StatisticsTests(unittest.TestCase):
    def test_average_rank_and_spearman_handle_ties(self):
        self.assertEqual(analyze.average_ranks([1, 1, 2, 2]), [1.5, 1.5, 3.5, 3.5])
        self.assertAlmostEqual(analyze.spearman([1, 1, 2, 2], [10, 10, 20, 20]), 1.0)

    def test_cluster_bootstrap_is_deterministic_and_uses_signed_swing(self):
        rows = [
            {"game_index": 1, "exposure": 1, "severe": True, "loss_swing_cp": 400},
            {"game_index": 1, "exposure": 0, "severe": False, "loss_swing_cp": -100},
            {"game_index": 2, "exposure": 1, "severe": False, "loss_swing_cp": -50},
            {"game_index": 2, "exposure": 0, "severe": False, "loss_swing_cp": -200},
        ]
        first = analyze.cluster_bootstrap(rows, "exposure", replications=50, seed=7)
        second = analyze.cluster_bootstrap(rows, "exposure", replications=50, seed=7)
        self.assertEqual(first, second)
        self.assertEqual(analyze._effect(rows, "exposure"), (0.5, 325.0))


if __name__ == "__main__":
    unittest.main()
